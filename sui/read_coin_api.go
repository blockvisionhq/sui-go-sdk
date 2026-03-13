// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package sui

import (
	"context"

	"github.com/block-vision/sui-go-sdk/common/httpconn"
	"github.com/block-vision/sui-go-sdk/models"
)

type IReadCoinFromSuiAPI interface {
	SuiXGetBalance(ctx context.Context, req models.SuiXGetBalanceRequest) (models.CoinBalanceResponse, error)
	SuiXGetAllBalance(ctx context.Context, req models.SuiXGetAllBalanceRequest) (models.CoinAllBalanceResponse, error)
	SuiXGetCoins(ctx context.Context, req models.SuiXGetCoinsRequest) (models.PaginatedCoinsResponse, error)
	SuiXGetAllCoins(ctx context.Context, req models.SuiXGetAllCoinsRequest) (models.PaginatedCoinsResponse, error)
	SuiXGetCoinMetadata(ctx context.Context, req models.SuiXGetCoinMetadataRequest) (models.CoinMetadataResponse, error)
	SuiXGetTotalSupply(ctx context.Context, req models.SuiXGetTotalSupplyRequest) (models.TotalSupplyResponse, error)
}

type suiReadCoinFromSuiImpl struct {
	handler *BaseRequestHandler
}

func newSuiReadCoinFromSuiImpl(conn *httpconn.HttpConn) *suiReadCoinFromSuiImpl {
	return &suiReadCoinFromSuiImpl{
		handler: NewBaseRequestHandler(conn),
	}
}

// SuiXGetBalance implements the method `suix_getBalance`, gets the total Coin balance for each coin type owned by an address.
func (s *suiReadCoinFromSuiImpl) SuiXGetBalance(ctx context.Context, req models.SuiXGetBalanceRequest) (models.CoinBalanceResponse, error) {
	var rsp models.CoinBalanceResponse
	params := []interface{}{req.Owner, req.CoinType}
	err := s.handler.ExecuteRequest(ctx, "suix_getBalance", params, &rsp)
	return rsp, err
}

// SuiXGetAllBalance implements the method `suix_getAllBalances`, gets all Coin balances owned by an address.
func (s *suiReadCoinFromSuiImpl) SuiXGetAllBalance(ctx context.Context, req models.SuiXGetAllBalanceRequest) (models.CoinAllBalanceResponse, error) {
	var rsp models.CoinAllBalanceResponse
	params := []interface{}{req.Owner}
	err := s.handler.ExecuteRequest(ctx, "suix_getAllBalances", params, &rsp)
	return rsp, err
}

// SuiXGetCoins implements the method `suix_getCoins`, gets a list of Coin objects by type owned by an address.
func (s *suiReadCoinFromSuiImpl) SuiXGetCoins(ctx context.Context, req models.SuiXGetCoinsRequest) (models.PaginatedCoinsResponse, error) {
	var rsp models.PaginatedCoinsResponse
	params := []interface{}{req.Owner, req.CoinType, req.Cursor, req.Limit}
	err := s.handler.ExecuteRequestWithValidation(ctx, "suix_getCoins", params, req, &rsp)
	return rsp, err
}

// SuiXGetAllCoins implements the method `suix_getAllCoins`, gets all Coin objects owned by an address.
func (s *suiReadCoinFromSuiImpl) SuiXGetAllCoins(ctx context.Context, req models.SuiXGetAllCoinsRequest) (models.PaginatedCoinsResponse, error) {
	var rsp models.PaginatedCoinsResponse
	params := []interface{}{req.Owner, req.Cursor, req.Limit}
	err := s.handler.ExecuteRequestWithValidation(ctx, "suix_getAllCoins", params, req, &rsp)
	return rsp, err
}

// SuiXGetCoinMetadata implements the method `suix_getCoinMetadata`, gets metadata(e.g., symbol, decimals) for a coin.
func (s *suiReadCoinFromSuiImpl) SuiXGetCoinMetadata(ctx context.Context, req models.SuiXGetCoinMetadataRequest) (models.CoinMetadataResponse, error) {
	var rsp models.CoinMetadataResponse
	params := []interface{}{req.CoinType}
	err := s.handler.ExecuteRequest(ctx, "suix_getCoinMetadata", params, &rsp)
	return rsp, err
}

// SuiXGetTotalSupply implements the method `suix_getTotalSupply`, gets total supply for a coin
func (s *suiReadCoinFromSuiImpl) SuiXGetTotalSupply(ctx context.Context, req models.SuiXGetTotalSupplyRequest) (models.TotalSupplyResponse, error) {
	var rsp models.TotalSupplyResponse
	params := []interface{}{req.CoinType}
	err := s.handler.ExecuteRequest(ctx, "suix_getTotalSupply", params, &rsp)
	return rsp, err
}
