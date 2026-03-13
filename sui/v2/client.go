// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v2

import (
	"context"

	"github.com/block-vision/sui-go-sdk/common/grpcconn"
	"github.com/block-vision/sui-go-sdk/common/httpconn"
	"github.com/block-vision/sui-go-sdk/sui/v2/grpc_client"
	"github.com/block-vision/sui-go-sdk/sui/v2/json_rpc_client"
)

// ClientOptions defines options for creating a v2 Sui client.
// Provide either GrpcClient (for gRPC) or HttpConn (for JSON RPC).
type ClientOptions struct {
	GrpcClient *grpcconn.SuiGrpcClient
	HttpConn   *httpconn.HttpConn
}

// Client is the unified v2 Sui client interface.
// It is implemented by both grpc_client.Client and json_rpc_client.Client.
type Client interface {
	GetChainIdentifier(ctx context.Context) (*GetChainIdentifierResponse, error)
	GetReferenceGasPrice(ctx context.Context) (*GetReferenceGasPriceResponse, error)
	GetCurrentSystemState(ctx context.Context) (*GetCurrentSystemStateResponse, error)
	GetBalance(ctx context.Context, options GetBalanceOptions) (*GetBalanceResponse, error)
	ListBalances(ctx context.Context, options ListBalancesOptions) (*ListBalancesResponse, error)
	ListCoins(ctx context.Context, options ListCoinsOptions) (*ListCoinsResponse, error)
	GetCoinMetadata(ctx context.Context, options GetCoinMetadataOptions) (*GetCoinMetadataResponse, error)
	GetObjects(ctx context.Context, options GetObjectsOptions) (*GetObjectsResponse, error)
	ListOwnedObjects(ctx context.Context, options ListOwnedObjectsOptions) (*ListOwnedObjectsResponse, error)
	GetTransaction(ctx context.Context, options GetTransactionOptions) (*TransactionResult, error)
	ExecuteTransaction(ctx context.Context, options ExecuteTransactionOptions) (*TransactionResult, error)
	SimulateTransaction(ctx context.Context, options SimulateTransactionOptions) (*SimulateTransactionResult, error)
	ListDynamicFields(ctx context.Context, options ListDynamicFieldsOptions) (*ListDynamicFieldsResponse, error)
	DefaultNameServiceName(ctx context.Context, options DefaultNameServiceNameOptions) (*DefaultNameServiceNameResponse, error)
	GetMoveFunction(ctx context.Context, options GetMoveFunctionOptions) (*GetMoveFunctionResponse, error)
	VerifyZkLoginSignature(ctx context.Context, options VerifyZkLoginSignatureOptions) (*ZkLoginVerifyResponse, error)
}

// NewClient creates a new v2 Sui client. Pass either GrpcClient or HttpConn.
func NewClient(opts ClientOptions) (Client, error) {
	if opts.GrpcClient != nil {
		return grpc_client.NewClient(grpc_client.ClientOptions{GrpcClient: opts.GrpcClient})
	}
	if opts.HttpConn != nil {
		return json_rpc_client.NewClient(json_rpc_client.ClientOptions{HttpConn: opts.HttpConn})
	}
	return nil, ErrClientRequired
}
