// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package json_rpc_client

import (
	"github.com/block-vision/sui-go-sdk/common/httpconn"
)

// ClientOptions defines options for creating a v2 Sui JSON RPC client
type ClientOptions struct {
	HttpConn *httpconn.HttpConn
}

// Client is the v2 Sui client that uses JSON RPC as the underlying transport.
type Client struct {
	conn *httpconn.HttpConn
}

// NewClient creates a new v2 Sui client with JSON RPC backend
func NewClient(opts ClientOptions) (*Client, error) {
	if opts.HttpConn == nil {
		return nil, ErrHttpConnRequired
	}

	return &Client{
		conn: opts.HttpConn,
	}, nil
}
