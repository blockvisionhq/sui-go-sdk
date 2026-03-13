// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package grpc_client

import (
	"context"
	"encoding/base64"
	"fmt"

	v2proto "github.com/block-vision/sui-go-sdk/pb/sui/rpc/v2"
	. "github.com/block-vision/sui-go-sdk/sui/v2/types"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ListCoins returns Coin<T> objects owned by the given address.
func (c *Client) ListCoins(ctx context.Context, options ListCoinsOptions) (*ListCoinsResponse, error) {
	stateService, err := c.grpcClient.StateService(ctx)
	if err != nil {
		return nil, wrapTransportError("ListCoins", "StateService", err)
	}

	coinType := SUI_TYPE_ARG
	if options.CoinType != nil {
		coinType = *options.CoinType
	}

	var pageToken []byte
	if options.Cursor != nil {
		pageToken, err = base64.StdEncoding.DecodeString(*options.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
	}

	objectType := fmt.Sprintf("0x2::coin::Coin<%s>", coinType)
	resp, err := stateService.ListOwnedObjects(ctx, &v2proto.ListOwnedObjectsRequest{
		Owner:      &options.Owner,
		ObjectType: &objectType,
		PageToken:  pageToken,
		PageSize:   options.Limit,
		ReadMask: &fieldmaskpb.FieldMask{
			Paths: []string{"owner", "object_type", "digest", "version", "object_id", "balance"},
		},
	})
	if err != nil {
		return nil, wrapTransportError("ListCoins", "ListOwnedObjects", err)
	}

	coins := make([]Coin, len(resp.Objects))
	for i, obj := range resp.Objects {
		coins[i] = Coin{
			ObjectId: obj.GetObjectId(),
			Version:  fmt.Sprintf("%d", obj.GetVersion()),
			Digest:   obj.GetDigest(),
			Owner:    convertGrpcOwner(obj.GetOwner()),
			Type:     coinType,
			Balance:  fmt.Sprintf("%d", obj.GetBalance()),
		}
	}

	var cursor *string
	if len(resp.NextPageToken) > 0 {
		s := base64.StdEncoding.EncodeToString(resp.NextPageToken)
		cursor = &s
	}

	return &ListCoinsResponse{
		Objects:     coins,
		Cursor:      cursor,
		HasNextPage: len(resp.NextPageToken) > 0,
	}, nil
}

// GetBalance returns the aggregated balance of a coin type for an address.
func (c *Client) GetBalance(ctx context.Context, options GetBalanceOptions) (*GetBalanceResponse, error) {
	stateService, err := c.grpcClient.StateService(ctx)
	if err != nil {
		return nil, wrapTransportError("GetBalance", "StateService", err)
	}

	coinType := SUI_TYPE_ARG
	if options.CoinType != nil {
		coinType = *options.CoinType
	}

	resp, err := stateService.GetBalance(ctx, &v2proto.GetBalanceRequest{
		Owner:    &options.Owner,
		CoinType: &coinType,
	})
	if err != nil {
		return nil, wrapTransportError("GetBalance", "GetBalance", err)
	}

	if resp.Balance == nil {
		return &GetBalanceResponse{
			Balance: Balance{
				CoinType:       coinType,
				Balance_:       "0",
				CoinBalance:    "0",
				AddressBalance: "0",
			},
		}, nil
	}

	bal := resp.Balance
	balStr := fmt.Sprintf("%d", bal.GetBalance())
	return &GetBalanceResponse{
		Balance: Balance{
			CoinType:       bal.GetCoinType(),
			Balance_:       balStr,
			CoinBalance:    balStr,
			AddressBalance: "0",
		},
	}, nil
}

// GetCoinMetadata returns on-chain metadata for the given coin type.
func (c *Client) GetCoinMetadata(ctx context.Context, options GetCoinMetadataOptions) (*GetCoinMetadataResponse, error) {
	stateService, err := c.grpcClient.StateService(ctx)
	if err != nil {
		return nil, wrapTransportError("GetCoinMetadata", "StateService", err)
	}

	resp, err := stateService.GetCoinInfo(ctx, &v2proto.GetCoinInfoRequest{
		CoinType: &options.CoinType,
	})
	if err != nil {
		return &GetCoinMetadataResponse{CoinMetadata: nil}, nil
	}

	if resp.Metadata == nil {
		return &GetCoinMetadataResponse{CoinMetadata: nil}, nil
	}

	m := resp.Metadata

	var idPtr *string
	if id := m.GetId(); id != "" {
		idPtr = &id
	}

	iconUrl := m.GetIconUrl()
	iconPtr := &iconUrl

	return &GetCoinMetadataResponse{
		CoinMetadata: &CoinMetadata{
			Id:          idPtr,
			Decimals:    int(m.GetDecimals()),
			Name:        m.GetName(),
			Symbol:      m.GetSymbol(),
			Description: m.GetDescription(),
			IconUrl:     iconPtr,
		},
	}, nil
}

// ListBalances returns all coin-type balances for an address.
func (c *Client) ListBalances(ctx context.Context, options ListBalancesOptions) (*ListBalancesResponse, error) {
	stateService, err := c.grpcClient.StateService(ctx)
	if err != nil {
		return nil, wrapTransportError("ListBalances", "StateService", err)
	}

	var pageToken []byte
	if options.Cursor != nil {
		var err error
		pageToken, err = base64.StdEncoding.DecodeString(*options.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
	}

	resp, err := stateService.ListBalances(ctx, &v2proto.ListBalancesRequest{
		Owner:     &options.Owner,
		PageToken: pageToken,
		PageSize:  options.Limit,
	})
	if err != nil {
		return nil, wrapTransportError("ListBalances", "ListBalances", err)
	}

	balances := make([]Balance, len(resp.Balances))
	for i, bal := range resp.Balances {
		balStr := fmt.Sprintf("%d", bal.GetBalance())
		balances[i] = Balance{
			CoinType:       bal.GetCoinType(),
			Balance_:       balStr,
			CoinBalance:    balStr,
			AddressBalance: "0",
		}
	}

	var cursor *string
	if len(resp.NextPageToken) > 0 {
		s := base64.StdEncoding.EncodeToString(resp.NextPageToken)
		cursor = &s
	}

	return &ListBalancesResponse{
		Balances:    balances,
		Cursor:      cursor,
		HasNextPage: len(resp.NextPageToken) > 0,
	}, nil
}
