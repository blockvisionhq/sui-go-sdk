// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package json_rpc_client

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/models/sui_json_rpc_types"
	"github.com/block-vision/sui-go-sdk/sui/v2/types"
)

// ListDynamicFields lists the dynamic fields of a parent object.
func (c *Client) ListDynamicFields(ctx context.Context, options types.ListDynamicFieldsOptions) (*types.ListDynamicFieldsResponse, error) {
	limit := uint64(50)
	if options.Limit != nil {
		limit = uint64(*options.Limit)
	}

	params := []interface{}{options.ParentId, options.Cursor, limit}

	var rsp models.PaginatedDynamicFieldInfoResponse
	if err := c.executeRequest(ctx, "suix_getDynamicFields", params, &rsp); err != nil {
		return nil, wrapTransportError("ListDynamicFields", "suix_getDynamicFields", err)
	}

	objectByID := make(map[string]types.Object)
	if len(rsp.Data) > 0 {
		objIds := make([]string, len(rsp.Data))
		for i, df := range rsp.Data {
			objIds[i] = df.ObjectId
		}
		var objRsps []*models.SuiObjectResponse
		opts := models.SuiObjectDataOptions{ShowType: true, ShowOwner: true, ShowDisplay: true, ShowContent: true}
		if err := c.executeRequest(ctx, "sui_multiGetObjects", []interface{}{objIds, opts}, &objRsps); err == nil {
			for _, or := range objRsps {
				if or == nil || or.Data == nil {
					continue
				}
				objectByID[or.Data.ObjectId] = convertSuiObjectData(or.Data, types.ObjectInclude{JSON: true})
			}
		}
	}

	fields := make([]types.DynamicFieldEntry, len(rsp.Data))
	for i, df := range rsp.Data {
		nameBcs, _ := base64.StdEncoding.DecodeString(df.BcsName)
		fieldObj := types.Object{
			ObjectId: df.ObjectId,
			Version:  strconv.FormatUint(df.Version, 10),
			Digest:   df.Digest,
		}
		if obj, ok := objectByID[df.ObjectId]; ok {
			fieldObj = obj
		}

		kind := convertDynamicFieldKind(df.Type)
		entry := types.DynamicFieldEntry{
			FieldId:     df.ObjectId,
			Type:        fieldObj.Type,
			ValueType:   normalizeCoinType(df.ObjectType),
			Name:        types.DynamicFieldName{Type: fieldObj.Type, BCS: nameBcs},
			Kind:        kind,
			FieldObject: fieldObj,
		}
		if kind == types.DynamicFieldKindDynamicObject {
			childID := df.ObjectId
			entry.ChildId = &childID
		}
		fields[i] = entry
	}

	var cursor *string
	if rsp.NextCursor != "" {
		cursor = &rsp.NextCursor
	}

	return &types.ListDynamicFieldsResponse{
		DynamicFields: fields,
		Cursor:        cursor,
		HasNextPage:   rsp.HasNextPage,
	}, nil
}

// DefaultNameServiceName resolves the primary .sui name for an address.
func (c *Client) DefaultNameServiceName(ctx context.Context, options types.DefaultNameServiceNameOptions) (*types.DefaultNameServiceNameResponse, error) {
	limit := uint64(1)
	params := []interface{}{options.Address, nil, limit}

	var nameRsp models.SuiXResolveNameServiceNamesResponse
	if err := c.executeRequest(ctx, "suix_resolveNameServiceNames", params, &nameRsp); err != nil {
		return &types.DefaultNameServiceNameResponse{
			Data: types.DefaultNameServiceNameData{
				Name:                nil,
				RegistrationNFTId:   "",
				ExpirationTimestamp: "",
				TargetAddress:       options.Address,
			},
		}, nil
	}

	var namePtr *string
	if len(nameRsp.Data) > 0 && nameRsp.Data[0] != "" {
		namePtr = &nameRsp.Data[0]
	}

	return &types.DefaultNameServiceNameResponse{
		Data: types.DefaultNameServiceNameData{
			Name:                namePtr,
			RegistrationNFTId:   "",
			ExpirationTimestamp: "",
			TargetAddress:       options.Address,
		},
	}, nil
}

// GetMoveFunction returns the descriptor of a Move function.
func (c *Client) GetMoveFunction(ctx context.Context, options types.GetMoveFunctionOptions) (*types.GetMoveFunctionResponse, error) {
	params := []interface{}{options.PackageId, options.ModuleName, options.Name}

	var rsp sui_json_rpc_types.SuiMoveNormalizedFunction
	if err := c.executeRequest(ctx, "sui_getNormalizedMoveFunction", params, &rsp); err != nil {
		return nil, wrapTransportError("GetMoveFunction", "sui_getNormalizedMoveFunction", err)
	}

	fn := types.FunctionResponse{
		PackageId:      options.PackageId,
		ModuleName:     options.ModuleName,
		Name:           options.Name,
		Visibility:     convertVisibility(rsp.Visibility),
		IsEntry:        rsp.IsEntry,
		TypeParameters: convertTypeParams(rsp),
		Parameters:     convertOpenSignatures(rsp.Parameters),
		Returns:        convertOpenSignatures(rsp.Return_),
	}

	return &types.GetMoveFunctionResponse{Function: fn}, nil
}

func convertVisibility(v interface{}) types.Visibility {
	if v == nil {
		return types.VisibilityUnknown
	}
	switch s := v.(type) {
	case string:
		switch s {
		case "Public", "public":
			return types.VisibilityPublic
		case "Friend", "friend":
			return types.VisibilityFriend
		case "Private", "private":
			return types.VisibilityPrivate
		}
	}
	return types.VisibilityUnknown
}

func convertTypeParams(fn sui_json_rpc_types.SuiMoveNormalizedFunction) []types.TypeParameter {
	return convertTypeParamsRaw(fn.TypeParameters)
}

func convertOpenSignatures(sigs []interface{}) []types.OpenSignature {
	if len(sigs) == 0 {
		return []types.OpenSignature{}
	}
	result := make([]types.OpenSignature, 0, len(sigs))
	for _, sig := range sigs {
		result = append(result, convertOpenSignature(sig))
	}
	return result
}

func convertTypeParamsRaw(raw []interface{}) []types.TypeParameter {
	result := make([]types.TypeParameter, 0, len(raw))
	for _, item := range raw {
		tp := types.TypeParameter{Constraints: []types.Ability{}}
		m, ok := item.(map[string]interface{})
		if !ok {
			result = append(result, tp)
			continue
		}
		if phantom, ok := m["isPhantom"].(bool); ok {
			tp.IsPhantom = phantom
		}
		if abilities, ok := m["abilities"].([]interface{}); ok {
			tp.Constraints = convertAbilities(abilities)
		} else if constraints, ok := m["constraints"].([]interface{}); ok {
			tp.Constraints = convertAbilities(constraints)
		}
		result = append(result, tp)
	}
	return result
}

func convertAbilities(raw []interface{}) []types.Ability {
	result := make([]types.Ability, 0, len(raw))
	for _, item := range raw {
		s, ok := item.(string)
		if !ok {
			continue
		}
		switch strings.ToLower(s) {
		case "copy":
			result = append(result, types.AbilityCopy)
		case "drop":
			result = append(result, types.AbilityDrop)
		case "store":
			result = append(result, types.AbilityStore)
		case "key":
			result = append(result, types.AbilityKey)
		default:
			result = append(result, types.AbilityUnknown)
		}
	}
	return result
}

func convertOpenSignature(sig interface{}) types.OpenSignature {
	switch v := sig.(type) {
	case map[string]interface{}:
		if inner, ok := v["Reference"]; ok {
			ref := types.ReferenceTypeImmutable
			return types.OpenSignature{Reference: &ref, Body: convertOpenSignatureBody(inner)}
		}
		if inner, ok := v["MutableReference"]; ok {
			ref := types.ReferenceTypeMutable
			return types.OpenSignature{Reference: &ref, Body: convertOpenSignatureBody(inner)}
		}
	}
	return types.OpenSignature{Body: convertOpenSignatureBody(sig)}
}

func convertOpenSignatureBody(sig interface{}) types.OpenSignatureBody {
	switch v := sig.(type) {
	case string:
		switch strings.ToUpper(v) {
		case "U8":
			return types.OpenSignatureBody{Kind: types.OpenSignatureBodyKindU8}
		case "U16":
			return types.OpenSignatureBody{Kind: types.OpenSignatureBodyKindU16}
		case "U32":
			return types.OpenSignatureBody{Kind: types.OpenSignatureBodyKindU32}
		case "U64":
			return types.OpenSignatureBody{Kind: types.OpenSignatureBodyKindU64}
		case "U128":
			return types.OpenSignatureBody{Kind: types.OpenSignatureBodyKindU128}
		case "U256":
			return types.OpenSignatureBody{Kind: types.OpenSignatureBodyKindU256}
		case "BOOL":
			return types.OpenSignatureBody{Kind: types.OpenSignatureBodyKindBool}
		case "ADDRESS":
			return types.OpenSignatureBody{Kind: types.OpenSignatureBodyKindAddress}
		default:
			return types.OpenSignatureBody{Kind: types.OpenSignatureBodyKindUnknown}
		}
	case map[string]interface{}:
		if inner, ok := v["Vector"]; ok {
			body := convertOpenSignatureBody(inner)
			return types.OpenSignatureBody{Kind: types.OpenSignatureBodyKindVector, Vector: &body}
		}
		if inner, ok := v["Struct"]; ok {
			return types.OpenSignatureBody{
				Kind:     types.OpenSignatureBodyKindDatatype,
				Datatype: convertDatatypeSignature(inner),
			}
		}
		if inner, ok := v["TypeParameter"]; ok {
			return types.OpenSignatureBody{
				Kind:               types.OpenSignatureBodyKindTypeParameter,
				TypeParameterIndex: toInt(inner),
			}
		}
		if inner, ok := v["Reference"]; ok {
			return convertOpenSignatureBody(inner)
		}
		if inner, ok := v["MutableReference"]; ok {
			return convertOpenSignatureBody(inner)
		}
	}
	return types.OpenSignatureBody{Kind: types.OpenSignatureBodyKindUnknown}
}

func convertDatatypeSignature(raw interface{}) *types.DatatypeSignature {
	m, ok := raw.(map[string]interface{})
	if !ok {
		return &types.DatatypeSignature{TypeParameters: []types.OpenSignatureBody{}}
	}
	typeName := normalizeCoinType(fmt.Sprintf("%s::%s::%s", toString(m["address"]), toString(m["module"]), toString(m["name"])))
	argsRaw, _ := m["typeArguments"].([]interface{})
	args := make([]types.OpenSignatureBody, 0, len(argsRaw))
	for _, arg := range argsRaw {
		args = append(args, convertOpenSignatureBody(arg))
	}
	return &types.DatatypeSignature{
		TypeName:       typeName,
		TypeParameters: args,
	}
}

func convertDynamicFieldKind(kind string) types.DynamicFieldKind {
	switch strings.ToLower(kind) {
	case "dynamicobject":
		return types.DynamicFieldKindDynamicObject
	default:
		return types.DynamicFieldKindDynamicField
	}
}

func toString(v interface{}) string {
	s, _ := v.(string)
	return s
}

func toInt(v interface{}) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int32:
		return int(x)
	case int64:
		return int(x)
	default:
		return 0
	}
}

// VerifyZkLoginSignature verifies a zkLogin signature.
// Note: JSON RPC does not have a direct equivalent; returns Success: false with error message.
func (c *Client) VerifyZkLoginSignature(ctx context.Context, options types.VerifyZkLoginSignatureOptions) (*types.ZkLoginVerifyResponse, error) {
	return &types.ZkLoginVerifyResponse{
		Success: false,
		Errors:  []string{"VerifyZkLoginSignature is not supported via JSON RPC; use gRPC client instead"},
	}, nil
}
