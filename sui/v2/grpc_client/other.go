// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package grpc_client

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	v2proto "github.com/block-vision/sui-go-sdk/pb/sui/rpc/v2"
	. "github.com/block-vision/sui-go-sdk/sui/v2/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ListDynamicFields lists the dynamic fields of a parent object.
func (c *Client) ListDynamicFields(ctx context.Context, options ListDynamicFieldsOptions) (*ListDynamicFieldsResponse, error) {
	stateService, err := c.grpcClient.StateService(ctx)
	if err != nil {
		return nil, wrapTransportError("ListDynamicFields", "StateService", err)
	}

	var pageToken []byte
	if options.Cursor != nil {
		pageToken, err = base64.StdEncoding.DecodeString(*options.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
	}

	resp, err := stateService.ListDynamicFields(ctx, &v2proto.ListDynamicFieldsRequest{
		Parent:    &options.ParentId,
		PageToken: pageToken,
		PageSize:  options.Limit,
		ReadMask: &fieldmaskpb.FieldMask{
			Paths: []string{"field_id", "name", "value_type", "kind", "child_id", "field_object"},
		},
	})
	if err != nil {
		return nil, wrapTransportError("ListDynamicFields", "ListDynamicFields", err)
	}

	fields := make([]DynamicFieldEntry, len(resp.DynamicFields))
	for i, f := range resp.DynamicFields {
		entry := DynamicFieldEntry{
			FieldId:     f.GetFieldId(),
			FieldObject: convertGrpcObject(f.GetFieldObject(), ObjectInclude{JSON: true}),
			ValueType:   f.GetValueType(),
			Type:        f.GetFieldObject().GetObjectType(),
			Kind:        convertDynamicFieldKind(f.GetKind()),
		}

		// BCS-encoded name
		if nameBcs := f.GetName(); nameBcs != nil {
			entry.Name = DynamicFieldName{
				BCS:  nameBcs.Value,
				Type: f.GetFieldObject().GetObjectType(),
			}
		}

		if childId := f.GetChildId(); childId != "" {
			entry.ChildId = &childId
		}

		fields[i] = entry
	}

	var cursor *string
	if len(resp.NextPageToken) > 0 {
		s := base64.StdEncoding.EncodeToString(resp.NextPageToken)
		cursor = &s
	}

	return &ListDynamicFieldsResponse{
		DynamicFields: fields,
		Cursor:        cursor,
		HasNextPage:   len(resp.NextPageToken) > 0,
	}, nil
}

func convertDynamicFieldKind(k v2proto.DynamicField_DynamicFieldKind) DynamicFieldKind {
	switch k {
	case v2proto.DynamicField_OBJECT:
		return DynamicFieldKindDynamicObject
	default:
		return DynamicFieldKindDynamicField
	}
}

// VerifyZkLoginSignature verifies a zkLogin signature.
func (c *Client) VerifyZkLoginSignature(ctx context.Context, options VerifyZkLoginSignatureOptions) (*ZkLoginVerifyResponse, error) {
	sigService, err := c.grpcClient.SignatureVerificationService(ctx)
	if err != nil {
		return nil, wrapTransportError("VerifyZkLoginSignature", "SignatureVerificationService", err)
	}

	msgBytes, err := base64.StdEncoding.DecodeString(options.Bytes)
	if err != nil {
		return nil, fmt.Errorf("invalid message bytes: %w", err)
	}

	sigBytes, err := base64.StdEncoding.DecodeString(options.Signature)
	if err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}

	intentScope := string(options.IntentScope)

	resp, err := sigService.VerifySignature(ctx, &v2proto.VerifySignatureRequest{
		Message: &v2proto.Bcs{
			Name:  &intentScope,
			Value: msgBytes,
		},
		Signature: &v2proto.UserSignature{
			Bcs: &v2proto.Bcs{Value: sigBytes},
		},
		Address: &options.Address,
		Jwks:    []*v2proto.ActiveJwk{},
	})
	if err != nil {
		return nil, wrapTransportError("VerifyZkLoginSignature", "VerifySignature", err)
	}

	var errs []string
	if reason := resp.GetReason(); reason != "" {
		errs = append(errs, reason)
	}

	return &ZkLoginVerifyResponse{
		Success: resp.GetIsValid(),
		Errors:  errs,
	}, nil
}

// DefaultNameServiceName resolves the primary .sui name for an address.
func (c *Client) DefaultNameServiceName(ctx context.Context, options DefaultNameServiceNameOptions) (*DefaultNameServiceNameResponse, error) {
	nameService, err := c.grpcClient.NameService(ctx)
	if err != nil {
		return nil, wrapTransportError("DefaultNameServiceName", "NameService", err)
	}

	resp, err := nameService.ReverseLookupName(ctx, &v2proto.ReverseLookupNameRequest{
		Address: &options.Address,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return &DefaultNameServiceNameResponse{
				Data: DefaultNameServiceNameData{
					Name:                nil,
					RegistrationNFTId:   "",
					ExpirationTimestamp: "",
					TargetAddress:       options.Address,
				},
			}, nil
		}
		return nil, wrapTransportError("DefaultNameServiceName", "ReverseLookupName", err)
	}

	record := resp.GetRecord()
	if record == nil {
		return &DefaultNameServiceNameResponse{
			Data: DefaultNameServiceNameData{
				Name:                nil,
				RegistrationNFTId:   "",
				ExpirationTimestamp: "",
				TargetAddress:       options.Address,
			},
		}, nil
	}

	var namePtr *string
	if name := record.GetName(); name != "" {
		namePtr = &name
	}

	return &DefaultNameServiceNameResponse{
		Data: DefaultNameServiceNameData{
			Name:                namePtr,
			RegistrationNFTId:   record.GetRegistrationNftId(),
			ExpirationTimestamp: strconv.FormatInt(record.GetExpirationTimestamp().AsTime().UnixMilli(), 10),
			TargetAddress:       record.GetTargetAddress(),
		},
	}, nil
}

// GetMoveFunction returns the descriptor of a Move function.
func (c *Client) GetMoveFunction(ctx context.Context, options GetMoveFunctionOptions) (*GetMoveFunctionResponse, error) {
	moveService, err := c.grpcClient.MovePackageService(ctx)
	if err != nil {
		return nil, wrapTransportError("GetMoveFunction", "MovePackageService", err)
	}

	resp, err := moveService.GetFunction(ctx, &v2proto.GetFunctionRequest{
		PackageId:  &options.PackageId,
		ModuleName: &options.ModuleName,
		Name:       &options.Name,
	})
	if err != nil {
		return nil, wrapTransportError("GetMoveFunction", "GetFunction", err)
	}

	fn := resp.GetFunction()
	if fn == nil {
		return nil, fmt.Errorf("function not found")
	}

	return &GetMoveFunctionResponse{
		Function: convertGrpcFunctionDescriptor(fn, options.PackageId, options.ModuleName),
	}, nil
}

// ─── Move type converters ─────────────────────────────────────────────────────

func convertGrpcFunctionDescriptor(fn *v2proto.FunctionDescriptor, packageId, moduleName string) FunctionResponse {
	return FunctionResponse{
		PackageId:      packageId,
		ModuleName:     moduleName,
		Name:           fn.GetName(),
		Visibility:     convertFunctionVisibility(fn.GetVisibility()),
		IsEntry:        fn.GetIsEntry(),
		TypeParameters: convertGrpcTypeParameters(fn.GetTypeParameters()),
		Parameters:     convertGrpcOpenSignatures(fn.GetParameters()),
		Returns:        convertGrpcOpenSignatures(fn.GetReturns()),
	}
}

func convertFunctionVisibility(v v2proto.FunctionDescriptor_Visibility) Visibility {
	switch v {
	case v2proto.FunctionDescriptor_PUBLIC:
		return VisibilityPublic
	case v2proto.FunctionDescriptor_PRIVATE:
		return VisibilityPrivate
	default:
		return VisibilityUnknown
	}
}

func convertGrpcTypeParameters(tps []*v2proto.TypeParameter) []TypeParameter {
	result := make([]TypeParameter, len(tps))
	for i, tp := range tps {
		result[i] = TypeParameter{
			IsPhantom:   tp.GetIsPhantom(),
			Constraints: convertGrpcAbilities(tp.GetConstraints()),
		}
	}
	return result
}

func convertGrpcAbilities(abilities []v2proto.Ability) []Ability {
	result := make([]Ability, len(abilities))
	for i, a := range abilities {
		result[i] = convertGrpcAbility(a)
	}
	return result
}

func convertGrpcAbility(a v2proto.Ability) Ability {
	switch a {
	case v2proto.Ability_COPY:
		return AbilityCopy
	case v2proto.Ability_DROP:
		return AbilityDrop
	case v2proto.Ability_STORE:
		return AbilityStore
	case v2proto.Ability_KEY:
		return AbilityKey
	default:
		return AbilityUnknown
	}
}

func convertGrpcOpenSignatures(sigs []*v2proto.OpenSignature) []OpenSignature {
	result := make([]OpenSignature, len(sigs))
	for i, sig := range sigs {
		result[i] = convertGrpcOpenSignature(sig)
	}
	return result
}

func convertGrpcOpenSignature(sig *v2proto.OpenSignature) OpenSignature {
	os := OpenSignature{}

	switch sig.GetReference() {
	case v2proto.OpenSignature_MUTABLE:
		ref := ReferenceTypeMutable
		os.Reference = &ref
	case v2proto.OpenSignature_IMMUTABLE:
		ref := ReferenceTypeImmutable
		os.Reference = &ref
	}

	if body := sig.GetBody(); body != nil {
		os.Body = convertGrpcOpenSignatureBody(body)
	}

	return os
}

func convertGrpcOpenSignatureBody(body *v2proto.OpenSignatureBody) OpenSignatureBody {
	b := OpenSignatureBody{}

	switch body.GetType() {
	case v2proto.OpenSignatureBody_U8:
		b.Kind = OpenSignatureBodyKindU8
	case v2proto.OpenSignatureBody_U16:
		b.Kind = OpenSignatureBodyKindU16
	case v2proto.OpenSignatureBody_U32:
		b.Kind = OpenSignatureBodyKindU32
	case v2proto.OpenSignatureBody_U64:
		b.Kind = OpenSignatureBodyKindU64
	case v2proto.OpenSignatureBody_U128:
		b.Kind = OpenSignatureBodyKindU128
	case v2proto.OpenSignatureBody_U256:
		b.Kind = OpenSignatureBodyKindU256
	case v2proto.OpenSignatureBody_BOOL:
		b.Kind = OpenSignatureBodyKindBool
	case v2proto.OpenSignatureBody_ADDRESS:
		b.Kind = OpenSignatureBodyKindAddress
	case v2proto.OpenSignatureBody_VECTOR:
		b.Kind = OpenSignatureBodyKindVector
		// Recurse: the element type is the first item in TypeParameterInstantiation.
		if items := body.GetTypeParameterInstantiation(); len(items) > 0 {
			elem := convertGrpcOpenSignatureBody(items[0])
			b.Vector = &elem
		}
	case v2proto.OpenSignatureBody_DATATYPE:
		b.Kind = OpenSignatureBodyKindDatatype
		typeArgs := body.GetTypeParameterInstantiation()
		args := make([]OpenSignatureBody, len(typeArgs))
		for i, arg := range typeArgs {
			args[i] = convertGrpcOpenSignatureBody(arg)
		}
		b.Datatype = &DatatypeSignature{
			TypeName:       body.GetTypeName(),
			TypeParameters: args,
		}
	case v2proto.OpenSignatureBody_TYPE_PARAMETER:
		b.Kind = OpenSignatureBodyKindTypeParameter
		b.TypeParameterIndex = int(body.GetTypeParameter())
	default:
		b.Kind = OpenSignatureBodyKindUnknown
	}

	return b
}
