// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package json_rpc_client

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/sui/v2/types"
)

// ListCoins returns Coin<T> objects owned by the given address.
func (c *Client) ListCoins(ctx context.Context, options types.ListCoinsOptions) (*types.ListCoinsResponse, error) {
	coinType := types.SUI_TYPE_ARG
	if options.CoinType != nil {
		coinType = *options.CoinType
	}

	limit := uint64(50)
	if options.Limit != nil {
		limit = uint64(*options.Limit)
	}

	var rsp models.PaginatedCoinsResponse
	params := []interface{}{options.Owner, coinType, options.Cursor, limit}
	if err := c.executeRequest(ctx, "suix_getCoins", params, &rsp); err != nil {
		return nil, wrapTransportError("ListCoins", "suix_getCoins", err)
	}

	coins := make([]types.Coin, len(rsp.Data))
	owner := types.ObjectOwner{Kind: types.ObjectOwnerKindAddress, AddressOwner: options.Owner}
	for i, d := range rsp.Data {
		coins[i] = types.Coin{
			ObjectId: d.CoinObjectId,
			Version:  d.Version,
			Digest:   d.Digest,
			Owner:    owner,
			Type:     d.CoinType,
			Balance:  d.Balance,
		}
	}

	var cursor *string
	if rsp.NextCursor != "" {
		cursor = &rsp.NextCursor
	}

	return &types.ListCoinsResponse{
		Objects:     coins,
		HasNextPage: rsp.HasNextPage,
		Cursor:      cursor,
	}, nil
}

// GetBalance returns the aggregated balance of a coin type for an address.
func (c *Client) GetBalance(ctx context.Context, options types.GetBalanceOptions) (*types.GetBalanceResponse, error) {
	coinType := types.SUI_TYPE_ARG
	if options.CoinType != nil {
		coinType = *options.CoinType
	}

	var rsp models.CoinBalanceResponse
	params := []interface{}{options.Owner, coinType}
	if err := c.executeRequest(ctx, "suix_getBalance", params, &rsp); err != nil {
		return nil, wrapTransportError("GetBalance", "suix_getBalance", err)
	}

	return &types.GetBalanceResponse{
		Balance: types.Balance{
			CoinType:       normalizeCoinType(rsp.CoinType),
			Balance_:       rsp.TotalBalance,
			CoinBalance:    rsp.TotalBalance,
			AddressBalance: "0",
		},
	}, nil
}

// GetCoinMetadata returns on-chain metadata for the given coin type.
func (c *Client) GetCoinMetadata(ctx context.Context, options types.GetCoinMetadataOptions) (*types.GetCoinMetadataResponse, error) {
	var rsp *models.CoinMetadataResponse
	params := []interface{}{options.CoinType}
	if err := c.executeRequest(ctx, "suix_getCoinMetadata", params, &rsp); err != nil {
		return &types.GetCoinMetadataResponse{CoinMetadata: nil}, nil
	}
	if rsp == nil {
		return &types.GetCoinMetadataResponse{CoinMetadata: nil}, nil
	}

	var idPtr *string
	if rsp.Id != "" {
		idPtr = &rsp.Id
	}
	iconUrl := rsp.IconUrl
	iconPtr := &iconUrl

	return &types.GetCoinMetadataResponse{
		CoinMetadata: &types.CoinMetadata{
			Id:          idPtr,
			Decimals:    rsp.Decimals,
			Name:        rsp.Name,
			Symbol:      rsp.Symbol,
			Description: rsp.Description,
			IconUrl:     iconPtr,
		},
	}, nil
}

// ListBalances returns all coin-type balances for an address.
func (c *Client) ListBalances(ctx context.Context, options types.ListBalancesOptions) (*types.ListBalancesResponse, error) {
	var rsp models.CoinAllBalanceResponse
	params := []interface{}{options.Owner}
	if err := c.executeRequest(ctx, "suix_getAllBalances", params, &rsp); err != nil {
		return nil, wrapTransportError("ListBalances", "suix_getAllBalances", err)
	}

	balances := make([]types.Balance, len(rsp))
	for i, b := range rsp {
		balances[i] = types.Balance{
			CoinType:       normalizeCoinType(b.CoinType),
			Balance_:       b.TotalBalance,
			CoinBalance:    b.TotalBalance,
			AddressBalance: "0",
		}
	}

	limit := 50
	if options.Limit != nil && *options.Limit > 0 {
		limit = int(*options.Limit)
	}

	start := 0
	if options.Cursor != nil && *options.Cursor != "" {
		raw, err := base64.StdEncoding.DecodeString(*options.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		start, err = strconv.Atoi(string(raw))
		if err != nil || start < 0 {
			return nil, fmt.Errorf("invalid cursor: %q", *options.Cursor)
		}
	}
	if start > len(balances) {
		start = len(balances)
	}

	end := start + limit
	if end > len(balances) {
		end = len(balances)
	}

	var cursor *string
	if end < len(balances) {
		next := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(end)))
		cursor = &next
	}

	return &types.ListBalancesResponse{
		Balances:    balances[start:end],
		HasNextPage: cursor != nil,
		Cursor:      cursor,
	}, nil
}
