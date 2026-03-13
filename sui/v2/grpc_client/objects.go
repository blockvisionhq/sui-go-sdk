// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package grpc_client

import (
	"context"
	"encoding/base64"
	"fmt"

	v2proto "github.com/block-vision/sui-go-sdk/pb/sui/rpc/v2"
	. "github.com/block-vision/sui-go-sdk/sui/v2/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// GetObjects retrieves multiple objects by their IDs.
func (c *Client) GetObjects(ctx context.Context, options GetObjectsOptions) (*GetObjectsResponse, error) {
	if len(options.ObjectIds) == 0 {
		return &GetObjectsResponse{Objects: []ObjectOrError{}}, nil
	}

	ledgerService, err := c.grpcClient.LedgerService(ctx)
	if err != nil {
		return nil, wrapTransportError("GetObjects", "LedgerService", err)
	}

	paths := objectReadMaskPaths(options.Include)

	// Process in batches of 50.
	const batchSize = 50
	var allObjects []ObjectOrError

	for i := 0; i < len(options.ObjectIds); i += batchSize {
		end := i + batchSize
		if end > len(options.ObjectIds) {
			end = len(options.ObjectIds)
		}

		batch := options.ObjectIds[i:end]
		requests := make([]*v2proto.GetObjectRequest, len(batch))
		for j, id := range batch {
			id := id
			requests[j] = &v2proto.GetObjectRequest{ObjectId: &id}
		}

		resp, err := ledgerService.BatchGetObjects(ctx, &v2proto.BatchGetObjectsRequest{
			Requests: requests,
			ReadMask: &fieldmaskpb.FieldMask{Paths: paths},
		})
		if err != nil {
			return nil, wrapTransportError("GetObjects", "BatchGetObjects", err)
		}

		for _, objResult := range resp.Objects {
			allObjects = append(allObjects, convertGrpcObjectResult(objResult, options.Include))
		}
	}

	return &GetObjectsResponse{Objects: allObjects}, nil
}

// ListOwnedObjects lists objects owned by the given address.
func (c *Client) ListOwnedObjects(ctx context.Context, options ListOwnedObjectsOptions) (*ListOwnedObjectsResponse, error) {
	stateService, err := c.grpcClient.StateService(ctx)
	if err != nil {
		return nil, wrapTransportError("ListOwnedObjects", "StateService", err)
	}

	paths := objectReadMaskPaths(options.Include)

	var pageToken []byte
	if options.Cursor != nil {
		pageToken, err = base64.StdEncoding.DecodeString(*options.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
	}

	req := &v2proto.ListOwnedObjectsRequest{
		Owner:     &options.Owner,
		PageToken: pageToken,
		PageSize:  options.Limit,
		ReadMask:  &fieldmaskpb.FieldMask{Paths: paths},
	}
	if options.Type != nil {
		req.ObjectType = options.Type
	}

	resp, err := stateService.ListOwnedObjects(ctx, req)
	if err != nil {
		return nil, wrapTransportError("ListOwnedObjects", "ListOwnedObjects", err)
	}

	objects := make([]Object, len(resp.Objects))
	for i, obj := range resp.Objects {
		objects[i] = convertGrpcObject(obj, options.Include)
	}

	var cursor *string
	if len(resp.NextPageToken) > 0 {
		s := base64.StdEncoding.EncodeToString(resp.NextPageToken)
		cursor = &s
	}

	return &ListOwnedObjectsResponse{
		Objects:     objects,
		Cursor:      cursor,
		HasNextPage: len(resp.NextPageToken) > 0,
	}, nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// objectReadMaskPaths builds the FieldMask path list from ObjectInclude.
func objectReadMaskPaths(include ObjectInclude) []string {
	paths := []string{"owner", "object_type", "digest", "version", "object_id"}
	if include.Content {
		paths = append(paths, "contents")
	}
	if include.PreviousTransaction {
		paths = append(paths, "previous_transaction")
	}
	if include.ObjectBcs {
		paths = append(paths, "bcs")
	}
	if include.JSON {
		paths = append(paths, "json")
	}
	return paths
}

// convertGrpcObjectResult converts a BatchGetObjects result entry to ObjectOrError.
func convertGrpcObjectResult(r *v2proto.GetObjectResult, include ObjectInclude) ObjectOrError {
	if r.Result == nil {
		errMsg := "object result is nil"
		return ObjectOrError{Error: &errMsg}
	}

	switch v := r.Result.(type) {
	case *v2proto.GetObjectResult_Error:
		msg := fmt.Sprintf("%s: %s",
			codes.Code(v.Error.GetCode()).String(),
			v.Error.GetMessage(),
		)
		return ObjectOrError{Error: &msg}
	case *v2proto.GetObjectResult_Object:
		obj := convertGrpcObject(v.Object, include)
		return ObjectOrError{Object: &obj}
	default:
		errMsg := "unexpected object result type"
		return ObjectOrError{Error: &errMsg}
	}
}

// convertGrpcObject converts a gRPC Object proto to the v2 Object type.
func convertGrpcObject(obj *v2proto.Object, include ObjectInclude) Object {
	o := Object{
		ObjectId: obj.GetObjectId(),
		Version:  fmt.Sprintf("%d", obj.GetVersion()),
		Digest:   obj.GetDigest(),
		Type:     obj.GetObjectType(),
		Owner:    convertGrpcOwner(obj.GetOwner()),
	}

	if include.PreviousTransaction {
		if prev := obj.GetPreviousTransaction(); prev != "" {
			o.PreviousTransaction = &prev
		}
	}

	if include.Content && obj.GetContents() != nil {
		o.Content = obj.GetContents().GetValue()
	}

	if include.ObjectBcs && obj.GetBcs() != nil {
		o.ObjectBcs = obj.GetBcs().GetValue()
	}

	if include.JSON && obj.GetJson() != nil {
		o.JSON = structpbValueToMap(obj.GetJson())
	}

	return o
}

// convertGrpcOwner converts a gRPC Owner proto to the v2 ObjectOwner type.
func convertGrpcOwner(owner *v2proto.Owner) ObjectOwner {
	if owner == nil {
		return ObjectOwner{Kind: ObjectOwnerKindUnknown}
	}

	switch owner.GetKind() {
	case v2proto.Owner_ADDRESS:
		return ObjectOwner{
			Kind:         ObjectOwnerKindAddress,
			AddressOwner: owner.GetAddress(),
		}
	case v2proto.Owner_OBJECT:
		return ObjectOwner{
			Kind:        ObjectOwnerKindObject,
			ObjectOwner: owner.GetAddress(),
		}
	case v2proto.Owner_IMMUTABLE:
		return ObjectOwner{Kind: ObjectOwnerKindImmutable}
	case v2proto.Owner_SHARED:
		return ObjectOwner{
			Kind: ObjectOwnerKindShared,
			Shared: &SharedOwnerPayload{
				InitialSharedVersion: fmt.Sprintf("%d", owner.GetVersion()),
			},
		}
	case v2proto.Owner_CONSENSUS_ADDRESS:
		return ObjectOwner{
			Kind: ObjectOwnerKindConsensusAddress,
			ConsensusAddress: &ConsensusOwnerPayload{
				StartVersion: fmt.Sprintf("%d", owner.GetVersion()),
				Owner:        owner.GetAddress(),
			},
		}
	}
	return ObjectOwner{Kind: ObjectOwnerKindUnknown}
}
