// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package grpc_client

import "github.com/block-vision/sui-go-sdk/sui/v2/types"

func wrapTransportError(method, operation string, cause error) error {
	return types.NewTransportError("grpc", method, operation, cause)
}

func wrapKnownError(base *types.SDKError, method, operation string, cause error) error {
	return types.WrapKnownError(base, "grpc", method, operation, cause)
}
