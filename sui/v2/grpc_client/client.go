// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package grpc_client

import (
	"github.com/block-vision/sui-go-sdk/common/grpcconn"
	"github.com/block-vision/sui-go-sdk/sui/v2/types"
)

// ClientOptions defines options for creating a v2 Sui client
type ClientOptions struct {
	GrpcClient *grpcconn.SuiGrpcClient
}

// Client is the v2 Sui client that uses gRPC as the underlying transport
type Client struct {
	grpcClient *grpcconn.SuiGrpcClient
}

// NewClient creates a new v2 Sui client with gRPC backend
func NewClient(opts ClientOptions) (*Client, error) {
	if opts.GrpcClient == nil {
		return nil, types.ErrGrpcClientRequired
	}

	return &Client{
		grpcClient: opts.GrpcClient,
	}, nil
}

// GetGrpcClient returns the underlying gRPC client
func (c *Client) GetGrpcClient() *grpcconn.SuiGrpcClient {
	return c.grpcClient
}
