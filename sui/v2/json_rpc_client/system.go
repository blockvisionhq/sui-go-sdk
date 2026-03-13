// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package json_rpc_client

import (
	"context"
	"strconv"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/sui/v2/types"
)

// GetReferenceGasPrice returns the current reference gas price.
func (c *Client) GetReferenceGasPrice(ctx context.Context) (*types.GetReferenceGasPriceResponse, error) {
	var gasPrice uint64
	if err := c.executeRequest(ctx, "suix_getReferenceGasPrice", []interface{}{}, &gasPrice); err != nil {
		return nil, wrapTransportError("GetReferenceGasPrice", "suix_getReferenceGasPrice", err)
	}
	return &types.GetReferenceGasPriceResponse{
		ReferenceGasPrice: strconv.FormatUint(gasPrice, 10),
	}, nil
}

// GetCurrentSystemState returns a snapshot of the current SuiSystemState.
func (c *Client) GetCurrentSystemState(ctx context.Context) (*types.GetCurrentSystemStateResponse, error) {
	var rsp models.SuiSystemStateSummary
	if err := c.executeRequest(ctx, "suix_getLatestSuiSystemState", []interface{}{}, &rsp); err != nil {
		return nil, wrapTransportError("GetCurrentSystemState", "suix_getLatestSuiSystemState", err)
	}

	info := types.SystemStateInfo{
		SystemStateVersion:              rsp.SystemStateVersion,
		Epoch:                           rsp.Epoch,
		ProtocolVersion:                 rsp.ProtocolVersion,
		ReferenceGasPrice:               rsp.ReferenceGasPrice,
		EpochStartTimestampMs:           rsp.EpochStartTimestampMs,
		SafeMode:                        rsp.SafeMode,
		SafeModeStorageRewards:          rsp.SafeModeStorageRewards,
		SafeModeComputationRewards:      rsp.SafeModeComputationRewards,
		SafeModeStorageRebates:          rsp.SafeModeStorageRebates,
		SafeModeNonRefundableStorageFee: rsp.SafeModeNonRefundableStorageFee,
		Parameters: types.SystemParameters{
			EpochDurationMs:              rsp.EpochDurationMs,
			StakeSubsidyStartEpoch:       rsp.StakeSubsidyStartEpoch,
			MaxValidatorCount:            rsp.MaxValidatorCount,
			MinValidatorJoiningStake:     rsp.MinValidatorJoiningStake,
			ValidatorLowStakeThreshold:   rsp.ValidatorLowStakeThreshold,
			ValidatorLowStakeGracePeriod: rsp.ValidatorLowStakeGracePeriod,
		},
		StorageFund: types.StorageFund{
			TotalObjectStorageRebates: rsp.StorageFundTotalObjectStorageRebates,
			NonRefundableBalance:      rsp.StorageFundNonRefundableBalance,
		},
		StakeSubsidy: types.StakeSubsidy{
			Balance:                   rsp.StakeSubsidyBalance,
			DistributionCounter:       rsp.StakeSubsidyDistributionCounter,
			CurrentDistributionAmount: rsp.StakeSubsidyCurrentDistributionAmount,
			StakeSubsidyPeriodLength:  rsp.StakeSubsidyPeriodLength,
			StakeSubsidyDecreaseRate:  rsp.StakeSubsidyDecreaseRate,
		},
	}

	return &types.GetCurrentSystemStateResponse{SystemState: info}, nil
}

// GetChainIdentifier returns the chain's unique identifier.
func (c *Client) GetChainIdentifier(ctx context.Context) (*types.GetChainIdentifierResponse, error) {
	var chainId string
	if err := c.executeRequest(ctx, "sui_getChainIdentifier", []interface{}{}, &chainId); err != nil {
		return nil, wrapTransportError("GetChainIdentifier", "sui_getChainIdentifier", err)
	}
	if chainId == "" {
		return nil, wrapKnownError(ErrChainIdentifierNotFound, "GetChainIdentifier", "empty chain identifier", nil)
	}
	return &types.GetChainIdentifierResponse{ChainIdentifier: chainId}, nil
}
