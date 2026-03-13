// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package json_rpc_client

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/sui/v2/types"
)

// GetObjects retrieves multiple objects by their IDs.
func (c *Client) GetObjects(ctx context.Context, options types.GetObjectsOptions) (*types.GetObjectsResponse, error) {
	if len(options.ObjectIds) == 0 {
		return &types.GetObjectsResponse{Objects: []types.ObjectOrError{}}, nil
	}

	opts := buildObjectOptions(options.Include)
	params := []interface{}{options.ObjectIds, opts}

	var rsp []*models.SuiObjectResponse
	if err := c.executeRequest(ctx, "sui_multiGetObjects", params, &rsp); err != nil {
		return nil, wrapTransportError("GetObjects", "sui_multiGetObjects", err)
	}

	objects := make([]types.ObjectOrError, len(rsp))
	for i, objRsp := range rsp {
		objects[i] = convertSuiObjectResponse(objRsp, options.Include)
	}

	return &types.GetObjectsResponse{Objects: objects}, nil
}

func buildObjectOptions(include types.ObjectInclude) models.SuiObjectDataOptions {
	return models.SuiObjectDataOptions{
		ShowType:                true,
		ShowContent:             include.Content || include.JSON, // JSON from Content.Fields, need content
		ShowBcs:                 include.ObjectBcs,
		ShowOwner:               true,
		ShowPreviousTransaction: include.PreviousTransaction,
		ShowDisplay:             include.JSON,
	}
}

func convertSuiObjectResponse(r *models.SuiObjectResponse, include types.ObjectInclude) types.ObjectOrError {
	if r == nil {
		errMsg := "object response is nil"
		return types.ObjectOrError{Error: &errMsg}
	}
	if r.Error != nil {
		errMsg := r.Error.Error
		return types.ObjectOrError{Error: &errMsg}
	}
	if r.Data == nil {
		errMsg := "object data is nil"
		return types.ObjectOrError{Error: &errMsg}
	}

	obj := convertSuiObjectData(r.Data, include)
	return types.ObjectOrError{Object: &obj}
}

func convertSuiObjectData(d *models.SuiObjectData, include types.ObjectInclude) types.Object {
	o := types.Object{
		ObjectId: d.ObjectId,
		Version:  d.Version,
		Digest:   d.Digest,
		Type:     normalizeCoinType(d.Type),
		Owner:    convertOwner(d.Owner),
	}

	if include.PreviousTransaction && d.PreviousTransaction != "" {
		o.PreviousTransaction = &d.PreviousTransaction
	}

	if include.Content && d.Bcs != nil {
		if data, err := base64.StdEncoding.DecodeString(d.Bcs.BcsBytes); err == nil {
			o.Content = data
		}
	}

	if include.ObjectBcs && d.Bcs != nil {
		if data, err := base64.StdEncoding.DecodeString(d.Bcs.BcsBytes); err == nil {
			o.ObjectBcs = data
		}
	}

	if include.JSON {
		// Content.Fields primary (key etc), Display as fallback
		if d.Content != nil && d.Content.Fields != nil {
			o.JSON = normalizeObjectJSON(d.Content.Fields, d.ObjectId)
		}
		if o.JSON == nil && d.Display.Data != nil {
			if m, ok := d.Display.Data.(map[string]interface{}); ok && len(m) > 0 {
				o.JSON = normalizeObjectJSON(m, d.ObjectId)
			}
		}
	}

	return o
}

func convertOwner(owner interface{}) types.ObjectOwner {
	if owner == nil {
		return types.ObjectOwner{Kind: types.ObjectOwnerKindUnknown}
	}
	// Owner can be ObjectOwner struct or map
	switch v := owner.(type) {
	case map[string]interface{}:
		if addr, ok := v["AddressOwner"].(string); ok {
			return types.ObjectOwner{Kind: types.ObjectOwnerKindAddress, AddressOwner: addr}
		}
		if addr, ok := v["ObjectOwner"].(string); ok {
			return types.ObjectOwner{Kind: types.ObjectOwnerKindObject, ObjectOwner: addr}
		}
		if shared, ok := v["Shared"].(map[string]interface{}); ok {
			ver, _ := shared["initial_shared_version"].(float64)
			return types.ObjectOwner{
				Kind: types.ObjectOwnerKindShared,
				Shared: &types.SharedOwnerPayload{
					InitialSharedVersion: fmt.Sprintf("%.0f", ver),
				},
			}
		}
	case models.ObjectOwner:
		o := types.ObjectOwner{}
		if v.AddressOwner != "" {
			o.Kind = types.ObjectOwnerKindAddress
			o.AddressOwner = v.AddressOwner
		} else if v.ObjectOwner != "" {
			o.Kind = types.ObjectOwnerKindObject
			o.ObjectOwner = v.ObjectOwner
		} else if v.Shared.InitialSharedVersion != 0 {
			o.Kind = types.ObjectOwnerKindShared
			o.Shared = &types.SharedOwnerPayload{
				InitialSharedVersion: fmt.Sprintf("%d", v.Shared.InitialSharedVersion),
			}
		} else {
			o.Kind = types.ObjectOwnerKindImmutable
		}
		return o
	}
	return types.ObjectOwner{Kind: types.ObjectOwnerKindUnknown}
}

// ListOwnedObjects lists objects owned by the given address.
func (c *Client) ListOwnedObjects(ctx context.Context, options types.ListOwnedObjectsOptions) (*types.ListOwnedObjectsResponse, error) {
	limit := uint64(50)
	if options.Limit != nil {
		limit = uint64(*options.Limit)
	}

	query := models.SuiObjectResponseQuery{
		Filter:  nil,
		Options: buildObjectOptions(options.Include),
	}
	if options.Type != nil {
		query.Filter = models.ObjectFilterByStructType{StructType: *options.Type}
	}

	params := []interface{}{options.Owner, query, options.Cursor, limit}

	var rsp models.PaginatedObjectsResponse
	if err := c.executeRequest(ctx, "suix_getOwnedObjects", params, &rsp); err != nil {
		return nil, wrapTransportError("ListOwnedObjects", "suix_getOwnedObjects", err)
	}

	var objects []types.Object
	for _, objRsp := range rsp.Data {
		if objRsp.Data != nil {
			objects = append(objects, convertSuiObjectData(objRsp.Data, options.Include))
		}
	}

	var cursor *string
	if rsp.NextCursor != "" {
		cursor = &rsp.NextCursor
	}

	return &types.ListOwnedObjectsResponse{
		Objects:     objects,
		HasNextPage: rsp.HasNextPage,
		Cursor:      cursor,
	}, nil
}
