// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package sui

import (
	"context"

	"github.com/block-vision/sui-go-sdk/common/httpconn"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/utils"
)

type IReadSystemFromSuiAPI interface {
	SuiGetCheckpoint(ctx context.Context, req models.SuiGetCheckpointRequest) (models.CheckpointResponse, error)
	SuiGetCheckpoints(ctx context.Context, req models.SuiGetCheckpointsRequest) (models.PaginatedCheckpointsResponse, error)
	SuiGetLatestCheckpointSequenceNumber(ctx context.Context) (uint64, error)
	SuiXGetReferenceGasPrice(ctx context.Context) (uint64, error)
	SuiXGetCommitteeInfo(ctx context.Context, req models.SuiXGetCommitteeInfoRequest) (models.SuiXGetCommitteeInfoResponse, error)
	SuiXGetStakes(ctx context.Context, req models.SuiXGetStakesRequest) ([]*models.DelegatedStakesResponse, error)
	SuiXGetStakesByIds(ctx context.Context, req models.SuiXGetStakesByIdsRequest) ([]*models.DelegatedStakesResponse, error)
	SuiXGetEpochs(ctx context.Context, req models.SuiXGetEpochsRequest) (models.PaginatedEpochInfoResponse, error)
	SuiXGetCurrentEpoch(ctx context.Context) (models.EpochInfo, error)
	SuiXGetLatestSuiSystemState(ctx context.Context) (models.SuiSystemStateSummary, error)
	SuiGetChainIdentifier(ctx context.Context) (string, error)
	SuiXGetValidatorsApy(ctx context.Context) (models.ValidatorsApy, error)
	SuiGetProtocolConfig(ctx context.Context, req models.SuiGetProtocolConfigRequest) (models.ProtocolConfigResponse, error)
}

type suiReadSystemFromSuiImpl struct {
	handler *BaseRequestHandler
}

func newSuiReadSystemFromSuiImpl(conn *httpconn.HttpConn) *suiReadSystemFromSuiImpl {
	return &suiReadSystemFromSuiImpl{
		handler: NewBaseRequestHandler(conn),
	}
}

// SuiGetCheckpoint implements the method `sui_getCheckpoint`, gets a checkpoint.
func (s *suiReadSystemFromSuiImpl) SuiGetCheckpoint(ctx context.Context, req models.SuiGetCheckpointRequest) (models.CheckpointResponse, error) {
	var rsp models.CheckpointResponse
	params := []interface{}{req.CheckpointID}
	err := s.handler.ExecuteRequest(ctx, "sui_getCheckpoint", params, &rsp)
	return rsp, err
}

// SuiGetCheckpoints implements the method `sui_getCheckpoints`, gets paginated list of checkpoints.
func (s *suiReadSystemFromSuiImpl) SuiGetCheckpoints(ctx context.Context, req models.SuiGetCheckpointsRequest) (models.PaginatedCheckpointsResponse, error) {
	var rsp models.PaginatedCheckpointsResponse
	params := []interface{}{req.Cursor, req.Limit, req.DescendingOrder}
	err := s.handler.ExecuteRequestWithValidation(ctx, "sui_getCheckpoints", params, req, &rsp)
	return rsp, err
}

// SuiGetLatestCheckpointSequenceNumber implements the method `sui_getLatestCheckpointSequenceNumber`, gets the sequence number of the latest checkpoint that has been executed.
func (s *suiReadSystemFromSuiImpl) SuiGetLatestCheckpointSequenceNumber(ctx context.Context) (uint64, error) {
	var rsp uint64
	err := s.handler.ExecuteRequest(ctx, "sui_getLatestCheckpointSequenceNumber", []interface{}{}, &rsp)
	return rsp, err
}

// SuiXGetReferenceGasPrice implements the method `suix_getReferenceGasPrice`, gets the reference gas price for the network.
func (s *suiReadSystemFromSuiImpl) SuiXGetReferenceGasPrice(ctx context.Context) (uint64, error) {
	var rsp uint64
	err := s.handler.ExecuteRequest(ctx, "suix_getReferenceGasPrice", []interface{}{}, &rsp)
	return rsp, err
}

// SuiXGetCommitteeInfo implements the method `suix_getCommitteeInfo`, gets the committee information for the asked `epoch`.
func (s *suiReadSystemFromSuiImpl) SuiXGetCommitteeInfo(ctx context.Context, req models.SuiXGetCommitteeInfoRequest) (models.SuiXGetCommitteeInfoResponse, error) {
	var rsp models.SuiXGetCommitteeInfoResponse
	params := []interface{}{req.Epoch}
	err := s.handler.ExecuteRequest(ctx, "suix_getCommitteeInfo", params, &rsp)
	return rsp, err
}

// SuiXGetStakes implements the method `suix_getStakes`, gets the delegated stakes for an address.
func (s *suiReadSystemFromSuiImpl) SuiXGetStakes(ctx context.Context, req models.SuiXGetStakesRequest) ([]*models.DelegatedStakesResponse, error) {
	var rsp []*models.DelegatedStakesResponse
	params := []interface{}{req.Owner}
	err := s.handler.ExecuteRequest(ctx, "suix_getStakes", params, &rsp)
	return rsp, err
}

// SuiXGetStakesByIds implements the method `suix_getStakesByIds`, return one or more delegated stake. If a Stake was withdrawn, its status will be Unstaked.
func (s *suiReadSystemFromSuiImpl) SuiXGetStakesByIds(ctx context.Context, req models.SuiXGetStakesByIdsRequest) ([]*models.DelegatedStakesResponse, error) {
	var rsp []*models.DelegatedStakesResponse
	params := []interface{}{req.StakedSuiIds}
	err := s.handler.ExecuteRequest(ctx, "suix_getStakesByIds", params, &rsp)
	return rsp, err
}

// SuiXGetEpochs implements the method `suix_getEpochs`, get a list of epoch info.
func (s *suiReadSystemFromSuiImpl) SuiXGetEpochs(ctx context.Context, req models.SuiXGetEpochsRequest) (models.PaginatedEpochInfoResponse, error) {
	var rsp models.PaginatedEpochInfoResponse
	params := []interface{}{req.Cursor, req.Limit, req.DescendingOrder}
	err := s.handler.ExecuteRequest(ctx, "suix_getEpochs", params, &rsp)
	return rsp, err
}

// SuiXGetCurrentEpoch implements the method `suix_getCurrentEpoch`, get current epoch info.
func (s *suiReadSystemFromSuiImpl) SuiXGetCurrentEpoch(ctx context.Context) (models.EpochInfo, error) {
	var rsp models.EpochInfo
	err := s.handler.ExecuteRequest(ctx, "suix_getCurrentEpoch", []interface{}{}, &rsp)
	return rsp, err
}

// SuiXGetLatestSuiSystemState implements the method `suix_getLatestSuiSystemState`, get the latest SUI system state object on-chain.
func (s *suiReadSystemFromSuiImpl) SuiXGetLatestSuiSystemState(ctx context.Context) (models.SuiSystemStateSummary, error) {
	var rsp models.SuiSystemStateSummary
	err := s.handler.ExecuteRequest(ctx, "suix_getLatestSuiSystemState", []interface{}{}, &rsp)
	return rsp, err
}

// SuiGetChainIdentifier implements the method `sui_getChainIdentifier`, return the chain's identifier.
func (s *suiReadSystemFromSuiImpl) SuiGetChainIdentifier(ctx context.Context) (string, error) {
	var rsp string
	err := s.handler.ExecuteRequest(ctx, "sui_getChainIdentifier", []interface{}{}, &rsp)
	return rsp, err
}

// SuiXGetValidatorsApy implements the method `suix_getValidatorsApy`, return the validator APY.
func (s *suiReadSystemFromSuiImpl) SuiXGetValidatorsApy(ctx context.Context) (models.ValidatorsApy, error) {
	var rsp models.ValidatorsApy
	err := s.handler.ExecuteRequest(ctx, "suix_getValidatorsApy", []interface{}{}, &rsp)
	return rsp, err
}

// SuiGetProtocolConfig implements the method `sui_getProtocolConfig`, return the protocol config table for the given version number.
// If the version number is not specified, If none is specified, the node uses the version of the latest epoch it has processed.
func (s *suiReadSystemFromSuiImpl) SuiGetProtocolConfig(ctx context.Context, req models.SuiGetProtocolConfigRequest) (models.ProtocolConfigResponse, error) {
	var rsp models.ProtocolConfigResponse
	params := make([]interface{}, 0)
	if utils.IsFieldNonEmpty(req, "Version") {
		params = append(params, req.Version)
	}
	err := s.handler.ExecuteRequest(ctx, "sui_getProtocolConfig", params, &rsp)
	return rsp, err
}
