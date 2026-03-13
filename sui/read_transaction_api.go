// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package sui

import (
	"context"

	"github.com/block-vision/sui-go-sdk/common/httpconn"
	"github.com/block-vision/sui-go-sdk/models"
)

type IReadTransactionFromSuiAPI interface {
	SuiGetTotalTransactionBlocks(ctx context.Context) (uint64, error)
	SuiGetTransactionBlock(ctx context.Context, req models.SuiGetTransactionBlockRequest) (models.SuiTransactionBlockResponse, error)
	SuiMultiGetTransactionBlocks(ctx context.Context, req models.SuiMultiGetTransactionBlocksRequest) (models.SuiMultiGetTransactionBlocksResponse, error)
	SuiXQueryTransactionBlocks(ctx context.Context, req models.SuiXQueryTransactionBlocksRequest) (models.SuiXQueryTransactionBlocksResponse, error)
	SuiDryRunTransactionBlock(ctx context.Context, req models.SuiDryRunTransactionBlockRequest) (models.SuiTransactionBlockResponse, error)
	SuiDevInspectTransactionBlock(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error)
}

type suiReadTransactionFromSuiImpl struct {
	handler *BaseRequestHandler
}

func newSuiReadTransactionFromSuiImpl(conn *httpconn.HttpConn) *suiReadTransactionFromSuiImpl {
	return &suiReadTransactionFromSuiImpl{
		handler: NewBaseRequestHandler(conn),
	}
}

// SuiGetTotalTransactionBlocks implements the method `sui_getTotalTransactionBlocks`, gets the total number of transactions known to the node.
func (s *suiReadTransactionFromSuiImpl) SuiGetTotalTransactionBlocks(ctx context.Context) (uint64, error) {
	var rsp uint64
	err := s.handler.ExecuteRequest(ctx, "sui_getTotalTransactionBlocks", []interface{}{}, &rsp)
	return rsp, err
}

// SuiGetTransactionBlock implements the method `sui_getTransactionBlock`, gets the transaction response object for a specified transaction digest.
func (s *suiReadTransactionFromSuiImpl) SuiGetTransactionBlock(ctx context.Context, req models.SuiGetTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
	var rsp models.SuiTransactionBlockResponse
	params := []interface{}{req.Digest, req.Options}
	err := s.handler.ExecuteRequest(ctx, "sui_getTransactionBlock", params, &rsp)
	return rsp, err
}

// SuiMultiGetTransactionBlocks implements the method `sui_multiGetTransactionBlocks`, gets an ordered list of transaction responses.
func (s *suiReadTransactionFromSuiImpl) SuiMultiGetTransactionBlocks(ctx context.Context, req models.SuiMultiGetTransactionBlocksRequest) (models.SuiMultiGetTransactionBlocksResponse, error) {
	var rsp models.SuiMultiGetTransactionBlocksResponse
	params := []interface{}{req.Digests, req.Options}
	err := s.handler.ExecuteRequest(ctx, "sui_multiGetTransactionBlocks", params, &rsp)
	return rsp, err
}

// SuiXQueryTransactionBlocks implements the method `suix_queryTransactionBlocks`, gets list of transactions for a specified query criteria.
func (s *suiReadTransactionFromSuiImpl) SuiXQueryTransactionBlocks(ctx context.Context, req models.SuiXQueryTransactionBlocksRequest) (models.SuiXQueryTransactionBlocksResponse, error) {
	var rsp models.SuiXQueryTransactionBlocksResponse
	params := []interface{}{req.SuiTransactionBlockResponseQuery, req.Cursor, req.Limit, req.DescendingOrder}
	err := s.handler.ExecuteRequestWithValidation(ctx, "suix_queryTransactionBlocks", params, req, &rsp)
	return rsp, err
}

// SuiDryRunTransactionBlock implements the method `sui_dryRunTransactionBlock`, gets transaction execution effects including the gas cost summary, while the effects are not committed to the chain.
func (s *suiReadTransactionFromSuiImpl) SuiDryRunTransactionBlock(ctx context.Context, req models.SuiDryRunTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
	var rsp models.SuiTransactionBlockResponse
	params := []interface{}{req.TxBytes}
	err := s.handler.ExecuteRequest(ctx, "sui_dryRunTransactionBlock", params, &rsp)
	return rsp, err
}

// SuiDevInspectTransactionBlock implements the method `sui_devInspectTransactionBlock`, runs the transaction in dev-inspect mode.
// Which allows for nearly any transaction (or Move call) with any arguments.
// Detailed results are provided, including both the transaction effects and any return values.
func (s *suiReadTransactionFromSuiImpl) SuiDevInspectTransactionBlock(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
	var rsp models.SuiTransactionBlockResponse
	params := []interface{}{req.Sender, req.TxBytes}
	if req.GasPrice != "" {
		params = append(params, req.GasPrice)
	}
	if req.Epoch != "" {
		params = append(params, req.Epoch)
	}
	err := s.handler.ExecuteRequest(ctx, "sui_devInspectTransactionBlock", params, &rsp)
	return rsp, err
}
