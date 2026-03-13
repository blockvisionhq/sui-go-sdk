// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package json_rpc_client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/block-vision/sui-go-sdk/common/httpconn"
	"github.com/tidwall/gjson"
)

func (c *Client) executeRequest(ctx context.Context, method string, params []interface{}, result interface{}) error {
	respBytes, err := c.conn.Request(ctx, httpconn.Operation{
		Method: method,
		Params: params,
	})
	if err != nil {
		return fmt.Errorf("%s request failed: %w", method, err)
	}

	parsed := gjson.ParseBytes(respBytes)
	if errVal := parsed.Get("error"); errVal.Exists() {
		return errors.New(errVal.String())
	}

	resultData := parsed.Get("result")
	if !resultData.Exists() {
		return fmt.Errorf("no result field in %s response", method)
	}

	if strPtr, ok := result.(*string); ok {
		*strPtr = resultData.String()
		return nil
	}

	if uint64Ptr, ok := result.(*uint64); ok {
		*uint64Ptr = resultData.Uint()
		return nil
	}

	jsonData := resultData.Raw
	if jsonData == "" {
		jsonData = resultData.String()
	}
	if jsonData == "" {
		return errors.New("empty result")
	}

	return json.Unmarshal([]byte(jsonData), result)
}
