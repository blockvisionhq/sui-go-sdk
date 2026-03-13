// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package grpc_client

import (
	"context"
	"fmt"

	v2proto "github.com/block-vision/sui-go-sdk/pb/sui/rpc/v2"
	. "github.com/block-vision/sui-go-sdk/sui/v2/types"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// GetReferenceGasPrice returns the current reference gas price.
func (c *Client) GetReferenceGasPrice(ctx context.Context) (*GetReferenceGasPriceResponse, error) {
	ledgerService, err := c.grpcClient.LedgerService(ctx)
	if err != nil {
		return nil, wrapTransportError("GetReferenceGasPrice", "ledger service", err)
	}

	resp, err := ledgerService.GetEpoch(ctx, &v2proto.GetEpochRequest{
		ReadMask: &fieldmaskpb.FieldMask{
			Paths: []string{"reference_gas_price"},
		},
	})
	if err != nil {
		return nil, wrapTransportError("GetReferenceGasPrice", "get epoch", err)
	}

	gasPrice := "0"
	if epoch := resp.GetEpoch(); epoch != nil {
		if p := epoch.GetReferenceGasPrice(); p != 0 {
			gasPrice = fmt.Sprintf("%d", p)
		}
	}

	return &GetReferenceGasPriceResponse{
		ReferenceGasPrice: gasPrice,
	}, nil
}

// GetCurrentSystemState returns a snapshot of the current SuiSystemState.
func (c *Client) GetCurrentSystemState(ctx context.Context) (*GetCurrentSystemStateResponse, error) {
	ledgerService, err := c.grpcClient.LedgerService(ctx)
	if err != nil {
		return nil, wrapTransportError("GetCurrentSystemState", "ledger service", err)
	}

	resp, err := ledgerService.GetEpoch(ctx, &v2proto.GetEpochRequest{
		ReadMask: &fieldmaskpb.FieldMask{
			Paths: []string{
				"system_state.version",
				"system_state.epoch",
				"system_state.protocol_version",
				"system_state.reference_gas_price",
				"system_state.epoch_start_timestamp_ms",
				"system_state.safe_mode",
				"system_state.safe_mode_storage_rewards",
				"system_state.safe_mode_computation_rewards",
				"system_state.safe_mode_storage_rebates",
				"system_state.safe_mode_non_refundable_storage_fee",
				"system_state.parameters",
				"system_state.storage_fund",
				"system_state.stake_subsidy",
			},
		},
	})
	if err != nil {
		return nil, wrapTransportError("GetCurrentSystemState", "get epoch", err)
	}

	epoch := resp.GetEpoch()
	if epoch == nil || epoch.GetSystemState() == nil {
		return nil, wrapKnownError(ErrSystemStateNotFound, "GetCurrentSystemState", "missing system state", nil)
	}

	ss := epoch.GetSystemState()
	info := SystemStateInfo{
		SystemStateVersion:              fmt.Sprintf("%d", ss.GetVersion()),
		Epoch:                           fmt.Sprintf("%d", ss.GetEpoch()),
		ProtocolVersion:                 fmt.Sprintf("%d", ss.GetProtocolVersion()),
		ReferenceGasPrice:               fmt.Sprintf("%d", ss.GetReferenceGasPrice()),
		EpochStartTimestampMs:           fmt.Sprintf("%d", ss.GetEpochStartTimestampMs()),
		SafeMode:                        ss.GetSafeMode(),
		SafeModeStorageRewards:          fmt.Sprintf("%d", ss.GetSafeModeStorageRewards()),
		SafeModeComputationRewards:      fmt.Sprintf("%d", ss.GetSafeModeComputationRewards()),
		SafeModeStorageRebates:          fmt.Sprintf("%d", ss.GetSafeModeStorageRebates()),
		SafeModeNonRefundableStorageFee: fmt.Sprintf("%d", ss.GetSafeModeNonRefundableStorageFee()),
	}

	if params := ss.GetParameters(); params != nil {
		info.Parameters = SystemParameters{
			EpochDurationMs:              fmt.Sprintf("%d", params.GetEpochDurationMs()),
			StakeSubsidyStartEpoch:       fmt.Sprintf("%d", params.GetStakeSubsidyStartEpoch()),
			MaxValidatorCount:            fmt.Sprintf("%d", params.GetMaxValidatorCount()),
			MinValidatorJoiningStake:     fmt.Sprintf("%d", params.GetMinValidatorJoiningStake()),
			ValidatorLowStakeThreshold:   fmt.Sprintf("%d", params.GetValidatorLowStakeThreshold()),
			ValidatorLowStakeGracePeriod: fmt.Sprintf("%d", params.GetValidatorLowStakeGracePeriod()),
		}
	}

	if sf := ss.GetStorageFund(); sf != nil {
		info.StorageFund = StorageFund{
			TotalObjectStorageRebates: fmt.Sprintf("%d", sf.GetTotalObjectStorageRebates()),
			NonRefundableBalance:      fmt.Sprintf("%d", sf.GetNonRefundableBalance()),
		}
	}

	if sub := ss.GetStakeSubsidy(); sub != nil {
		info.StakeSubsidy = StakeSubsidy{
			Balance:                   fmt.Sprintf("%d", sub.GetBalance()),
			DistributionCounter:       fmt.Sprintf("%d", sub.GetDistributionCounter()),
			CurrentDistributionAmount: fmt.Sprintf("%d", sub.GetCurrentDistributionAmount()),
			StakeSubsidyPeriodLength:  fmt.Sprintf("%d", sub.GetStakeSubsidyPeriodLength()),
			StakeSubsidyDecreaseRate:  int(sub.GetStakeSubsidyDecreaseRate()),
		}
	}

	return &GetCurrentSystemStateResponse{SystemState: info}, nil
}

// GetChainIdentifier returns the chain's unique identifier (genesis checkpoint digest, base58).
func (c *Client) GetChainIdentifier(ctx context.Context) (*GetChainIdentifierResponse, error) {
	ledgerService, err := c.grpcClient.LedgerService(ctx)
	if err != nil {
		return nil, wrapTransportError("GetChainIdentifier", "ledger service", err)
	}

	resp, err := ledgerService.GetServiceInfo(ctx, &v2proto.GetServiceInfoRequest{})
	if err != nil {
		return nil, wrapTransportError("GetChainIdentifier", "get service info", err)
	}

	chainId := resp.GetChainId()
	if chainId == "" {
		return nil, wrapKnownError(ErrChainIdentifierNotFound, "GetChainIdentifier", "empty chain identifier", nil)
	}

	return &GetChainIdentifierResponse{ChainIdentifier: chainId}, nil
}
