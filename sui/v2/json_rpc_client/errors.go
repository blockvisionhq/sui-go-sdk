// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package json_rpc_client

import "github.com/block-vision/sui-go-sdk/sui/v2/types"

var (
	ErrHttpConnRequired        = types.ErrHttpConnRequired
	ErrTransactionNotFound     = types.ErrTransactionNotFound
	ErrChainIdentifierNotFound = types.ErrChainIdentifierNotFound
)

func wrapTransportError(method, operation string, cause error) error {
	return types.NewTransportError("json-rpc", method, operation, cause)
}

func wrapKnownError(base *types.SDKError, method, operation string, cause error) error {
	return types.WrapKnownError(base, "json-rpc", method, operation, cause)
}
