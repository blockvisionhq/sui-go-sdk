// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package json_rpc_client

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/mystenbcs"
	"github.com/block-vision/sui-go-sdk/sui/v2/types"
)

// GetTransaction fetches a transaction by digest.
func (c *Client) GetTransaction(ctx context.Context, options types.GetTransactionOptions) (*types.TransactionResult, error) {
	// Enable all options to get full RPC data (inputDigest, inputOwner, events, gasObject, etc.)
	opts := models.SuiTransactionBlockOptions{
		ShowInput:          true,
		ShowRawInput:       true,
		ShowEffects:        true,
		ShowEvents:         true,
		ShowBalanceChanges: true,
		ShowObjectChanges:  true,
	}

	params := []interface{}{options.Digest, opts}

	var rsp *models.SuiTransactionBlockResponse
	if err := c.executeRequest(ctx, "sui_getTransactionBlock", params, &rsp); err != nil {
		return nil, wrapTransportError("GetTransaction", "sui_getTransactionBlock", err)
	}
	if rsp == nil {
		return nil, wrapKnownError(ErrTransactionNotFound, "GetTransaction", "missing transaction "+options.Digest, nil)
	}

	tx := convertTxResponse(rsp, options.Include)
	return &types.TransactionResult{Transaction: tx}, nil
}

// ExecuteTransaction submits a signed transaction for execution.
func (c *Client) ExecuteTransaction(ctx context.Context, options types.ExecuteTransactionOptions) (*types.TransactionResult, error) {
	txB64 := base64.StdEncoding.EncodeToString(options.Transaction)
	params := []interface{}{
		txB64,
		options.Signatures,
		map[string]interface{}{
			"showInput":          options.Include.Transaction,
			"showRawInput":       options.Include.Bcs,
			"showEffects":        options.Include.Effects,
			"showEvents":         options.Include.Events,
			"showBalanceChanges": options.Include.BalanceChanges,
			"showObjectChanges":  options.Include.Effects || options.Include.ObjectTypes,
		},
		"WaitForEffectsCert",
	}

	var rsp *models.SuiTransactionBlockResponse
	if err := c.executeRequest(ctx, "sui_executeTransactionBlock", params, &rsp); err != nil {
		return nil, wrapTransportError("ExecuteTransaction", "sui_executeTransactionBlock", err)
	}
	if rsp == nil {
		return nil, wrapKnownError(ErrTransactionNotFound, "ExecuteTransaction", "missing transaction response", nil)
	}

	tx := convertTxResponse(rsp, options.Include)
	return &types.TransactionResult{Transaction: tx}, nil
}

// SimulateTransaction simulates a transaction without submitting it.
func (c *Client) SimulateTransaction(ctx context.Context, options types.SimulateTransactionOptions) (*types.SimulateTransactionResult, error) {
	txB64 := base64.StdEncoding.EncodeToString(options.Transaction)
	params := []interface{}{txB64,
		map[string]interface{}{
			"showInput":          options.Include.Transaction,
			"showRawInput":       options.Include.Bcs,
			"showEffects":        options.Include.Effects,
			"showEvents":         options.Include.Events,
			"showBalanceChanges": options.Include.BalanceChanges,
			"showObjectChanges":  options.Include.Effects || options.Include.ObjectTypes,
		},
	}

	var rsp *models.SuiTransactionBlockResponse
	if err := c.executeRequest(ctx, "sui_dryRunTransactionBlock", params, &rsp); err != nil {
		return nil, wrapTransportError("SimulateTransaction", "sui_dryRunTransactionBlock", err)
	}
	if rsp == nil {
		return nil, wrapKnownError(ErrTransactionNotFound, "SimulateTransaction", "missing transaction response", nil)
	}

	tx := convertTxResponse(rsp, options.Include.TransactionInclude)
	return &types.SimulateTransactionResult{Transaction: tx}, nil
}

func convertTxResponse(rsp *models.SuiTransactionBlockResponse, include types.TransactionInclude) *types.Transaction {
	tx := &types.Transaction{
		Digest:     rsp.Digest,
		Signatures: []string{},
	}

	if rsp.TimestampMs != "" {
		if ts, err := strconv.ParseInt(rsp.TimestampMs, 10, 64); err == nil {
			tx.Timestamp = ts
		}
	}
	if rsp.Checkpoint != "" {
		if cp, err := strconv.ParseUint(rsp.Checkpoint, 10, 64); err == nil {
			tx.Checkpoint = cp
		}
	}

	if rsp.Transaction.TxSignatures != nil {
		tx.Signatures = rsp.Transaction.TxSignatures
	}

	if rsp.Effects.Status.Status != "" || rsp.Effects.Status.Error != "" {
		st := rsp.Effects.Status
		if st.Status == "success" {
			tx.Status = types.ExecutionStatus{Success: true}
		} else {
			tx.Status = types.ExecutionStatus{
				Success: false,
				Error:   &types.ExecutionError{Message: st.Error, Kind: types.ExecutionErrorKindUnknown},
			}
		}
	}

	if rsp.Effects.ExecutedEpoch != "" {
		if ep, err := strconv.ParseUint(rsp.Effects.ExecutedEpoch, 10, 64); err == nil {
			tx.Epoch = ep
		}
	}

	if include.Effects {
		gu := rsp.Effects.GasUsed
		gasData := gasDataFromRsp(rsp)
		changedObjs := objectChangesToChangedObjects(rsp.ObjectChanges, &rsp.Effects, gasData)
		if rsp.Effects.TransactionDigest != "" || len(rsp.Effects.Dependencies) > 0 ||
			gu.ComputationCost != "" || gu.StorageCost != "" || gu.StorageRebate != "" ||
			rsp.Effects.Status.Status != "" || rsp.Effects.ExecutedEpoch != "" || len(changedObjs) > 0 {
			nonRefundable := gu.NonRefundableStorageFee
			if nonRefundable == "" {
				nonRefundable = "0"
			}
			tx.Effects = &types.TransactionEffects{
				TransactionDigest: rsp.Effects.TransactionDigest,
				Dependencies:      rsp.Effects.Dependencies,
				ChangedObjects:    changedObjs,
				GasObject:         gasObjectFromChangedObjects(changedObjs, gasData),
				GasUsed: types.GasCostSummary{
					ComputationCost:         gu.ComputationCost,
					StorageCost:             gu.StorageCost,
					StorageRebate:           gu.StorageRebate,
					NonRefundableStorageFee: nonRefundable,
				},
			}
			if rsp.Effects.ExecutedEpoch != "" {
				if ep, err := strconv.ParseUint(rsp.Effects.ExecutedEpoch, 10, 64); err == nil {
					tx.Effects.ExecutedEpoch = ep
				}
			}
			if rsp.Effects.Status.Status == "success" {
				tx.Effects.Status = types.ExecutionStatus{Success: true}
			} else {
				tx.Effects.Status = types.ExecutionStatus{
					Success: false,
					Error:   &types.ExecutionError{Message: rsp.Effects.Status.Error, Kind: types.ExecutionErrorKindUnknown},
				}
			}
		} else {
			co := objectChangesToChangedObjects(rsp.ObjectChanges, &rsp.Effects, gasData)
			tx.Effects = &types.TransactionEffects{
				TransactionDigest: rsp.Effects.TransactionDigest,
				Dependencies:      rsp.Effects.Dependencies,
				ChangedObjects:    co,
				GasObject:         gasObjectFromChangedObjects(co, gasData),
			}
			if rsp.Effects.Status.Status == "success" {
				tx.Effects.Status = types.ExecutionStatus{Success: true}
			} else {
				tx.Effects.Status = types.ExecutionStatus{
					Success: false,
					Error:   &types.ExecutionError{Message: rsp.Effects.Status.Error, Kind: types.ExecutionErrorKindUnknown},
				}
			}
		}
	}

	if include.ObjectTypes && len(rsp.ObjectChanges) > 0 {
		tx.ObjectTypes = make(map[string]string)
		for _, oc := range rsp.ObjectChanges {
			if oc.ObjectId != "" && oc.ObjectType != "" {
				tx.ObjectTypes[oc.ObjectId] = normalizeObjectTypeForGrpc(normalizeCoinType(oc.ObjectType))
			}
		}
	}

	if include.BalanceChanges && len(rsp.BalanceChanges) > 0 {
		tx.BalanceChanges = make([]types.BalanceChange, len(rsp.BalanceChanges))
		for i, bc := range rsp.BalanceChanges {
			tx.BalanceChanges[i] = types.BalanceChange{
				CoinType: normalizeCoinType(bc.CoinType),
				Address:  bc.GetBalanceChangeOwner(),
				Amount:   bc.Amount,
			}
		}
	}

	if include.Transaction && rsp.Transaction.Data.Sender != "" {
		data := &rsp.Transaction.Data
		gd := data.GasData
		payment := make([]types.ObjectRef, 0, len(gd.Payment))
		for _, p := range gd.Payment {
			payment = append(payment, types.ObjectRef{
				ObjectId: p.ObjectId,
				Version:  strconv.FormatUint(p.Version, 10),
				Digest:   p.Digest,
			})
		}
		txData := &types.TransactionData{
			Sender: data.Sender,
			Kind:   "PROGRAMMABLE_TRANSACTION",
			GasData: &types.GasData{
				Budget:  gd.Budget,
				Owner:   gd.Owner,
				Price:   gd.Price,
				Payment: payment,
			},
		}
		txData.Commands = convertTransactionsToCommands(data.Transaction.Transactions)
		txData.Inputs = convertSuiCallArgsToInputs(data.Transaction.Inputs)
		txData.Expiration = types.NoExpiration{}
		tx.TransactionData = txData
	}

	if include.Bcs && rsp.RawTransaction != "" {
		if data, err := base64.StdEncoding.DecodeString(rsp.RawTransaction); err == nil {
			tx.BCS = data
		}
	}

	return tx
}

func gasDataFromRsp(rsp *models.SuiTransactionBlockResponse) *models.SuiGasData {
	if rsp != nil && rsp.Transaction.Data.Sender != "" {
		return &rsp.Transaction.Data.GasData
	}
	return nil
}

// gasObjectFromChangedObjects finds gas coin from changedObjects (objectId in gasData.Payment) and returns it
func gasObjectFromChangedObjects(changedObjs []types.ChangedObject, gasData *models.SuiGasData) *types.ChangedObject {
	if gasData == nil || len(gasData.Payment) == 0 {
		return nil
	}
	gasIds := make(map[string]bool)
	for _, p := range gasData.Payment {
		if p.ObjectId != "" {
			gasIds[p.ObjectId] = true
		}
	}
	for i := range changedObjs {
		if gasIds[changedObjs[i].ObjectId] {
			return &changedObjs[i]
		}
	}
	return nil
}

func objectChangesToChangedObjects(changes []models.ObjectChange, effects *models.SuiEffects, gasData *models.SuiGasData) []types.ChangedObject {
	if len(changes) == 0 {
		return nil
	}
	// Build inputDigest map: SharedObjects and GasData.Payment
	inputDigestMap := make(map[string]string)
	if effects != nil {
		for _, so := range effects.SharedObjects {
			if so.ObjectId != "" && so.Digest != "" {
				inputDigestMap[so.ObjectId] = so.Digest
			}
		}
	}
	if gasData != nil {
		for _, p := range gasData.Payment {
			if p.ObjectId != "" && p.Digest != "" {
				inputDigestMap[p.ObjectId] = p.Digest
			}
		}
	}

	out := make([]types.ChangedObject, 0, len(changes))
	for _, oc := range changes {
		co := objectChangeToChangedObject(oc, inputDigestMap)
		if co.ObjectId != "" {
			out = append(out, co)
		}
	}
	return out
}

func objectChangeToChangedObject(oc models.ObjectChange, inputDigestMap map[string]string) types.ChangedObject {
	var inputState types.ChangedObjectInputState
	var outputState types.ChangedObjectOutputState
	var idOp types.ChangedObjectIdOperation

	switch oc.Type {
	case "created", "published":
		inputState = types.ChangedObjectInputDoesNotExist
		outputState = types.ChangedObjectOutputObjectWrite
		idOp = types.ChangedObjectIdOpCreated
	case "mutated", "transferred", "wrapped":
		inputState = types.ChangedObjectInputExists
		outputState = types.ChangedObjectOutputObjectWrite
		idOp = types.ChangedObjectIdOpNone
	case "deleted":
		inputState = types.ChangedObjectInputExists
		outputState = types.ChangedObjectOutputDoesNotExist
		idOp = types.ChangedObjectIdOpDeleted
	default:
		inputState = types.ChangedObjectInputUnknown
		outputState = types.ChangedObjectOutputUnknown
		idOp = types.ChangedObjectIdOpUnknown
	}

	outputOwner := convertObjectChangeOwner(oc.Owner)
	prevVer := oc.PreviousVersion
	ver := oc.Version

	co := types.ChangedObject{
		ObjectId:      oc.ObjectId,
		InputState:    inputState,
		OutputState:   outputState,
		IdOperation:   idOp,
		OutputOwner:   outputOwner,
		OutputDigest:  strPtr(oc.Digest),
		OutputVersion: strPtr(ver),
	}
	if prevVer != "" {
		co.InputVersion = &prevVer
	}

	// Supplement inputDigest/inputOwner from JSON-RPC data
	if inputDigest := inputDigestMap[oc.ObjectId]; inputDigest != "" {
		co.InputDigest = &inputDigest
		// Shared and gas objects: inputOwner same as outputOwner
		co.InputOwner = outputOwner
	} else if oc.Type == "transferred" && oc.Sender != "" {
		// Transferred: input from sender
		co.InputOwner = &types.ObjectOwner{Kind: types.ObjectOwnerKindAddress, AddressOwner: oc.Sender}
	} else if outputOwner != nil && (oc.Type == "mutated" || oc.Type == "wrapped") {
		// Mutate/wrapped: input owner same as output (object not transferred)
		co.InputOwner = outputOwner
	}
	// Note: owned object inputDigest not in JSON-RPC response, cannot supplement
	return co
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func convertTransactionsToCommands(txs []any) []map[string]types.Command {
	if len(txs) == 0 {
		return nil
	}
	out := make([]map[string]types.Command, 0, len(txs))
	for _, t := range txs {
		if m, ok := t.(map[string]interface{}); ok && len(m) > 0 {
			for k, v := range m {
				normalized := normalizeCommandForGrpc(v)
				out = append(out, map[string]types.Command{k: types.RawCommand{Data: normalized}})
				break
			}
		}
	}
	return out
}

func convertSuiCallArgsToInputs(args []models.SuiCallArg) []types.Input {
	if len(args) == 0 {
		return nil
	}
	out := make([]types.Input, 0, len(args))
	for _, a := range args {
		normalized := normalizeInputForGrpc(a)
		out = append(out, types.RawInput{Data: normalized})
	}
	return out
}

// normalizeObjectTypeForGrpc removes space after comma in type params to match gRPC (Pool<A, B> -> Pool<A,B>)
func normalizeObjectTypeForGrpc(s string) string {
	return strings.ReplaceAll(s, ", ", ",")
}

// normalizeCommandForGrpc converts JSON-RPC command to gRPC format: arguments (Input/Result/NestedResult -> index/result/subresult), type_arguments -> typeArguments
func normalizeCommandForGrpc(v interface{}) interface{} {
	m, ok := v.(map[string]interface{})
	if !ok {
		return v
	}
	out := make(map[string]interface{})
	for k, val := range m {
		if k == "type_arguments" {
			out["typeArguments"] = val
			continue
		}
		if k == "arguments" {
			if arr, ok := val.([]interface{}); ok {
				out["arguments"] = normalizeArgumentsForGrpc(arr)
				continue
			}
		}
		out[k] = val
	}
	return out
}

func normalizeArgumentsForGrpc(args []interface{}) []interface{} {
	out := make([]interface{}, 0, len(args))
	for _, a := range args {
		arg, ok := a.(map[string]interface{})
		if !ok {
			out = append(out, a)
			continue
		}
		norm := make(map[string]interface{})
		if nested, ok := arg["NestedResult"].([]interface{}); ok && len(nested) >= 2 {
			// NestedResult: [cmdIndex, subresultIndex] -> {result, subresult}
			if res, ok := getInt(nested[0]); ok {
				norm["result"] = res
			}
			if sub, ok := getInt(nested[1]); ok && sub != 0 {
				// gRPC omitempty: subresult 0 not output
				norm["subresult"] = sub
			}
		} else if idx, ok := getInt(arg["Input"]); ok {
			norm["index"] = idx
		} else if res, ok := getInt(arg["Result"]); ok {
			norm["result"] = res
		} else {
			norm = arg
		}
		out = append(out, norm)
	}
	return out
}

func getInt(v interface{}) (int, bool) {
	switch x := v.(type) {
	case int:
		return x, true
	case int64:
		return int(x), true
	case float64:
		return int(x), true
	default:
		return 0, false
	}
}

// normalizeInputForGrpc converts JSON-RPC input to gRPC format: shared keeps 3 fields; pure to bytes; owned keeps objectId/version/digest
func normalizeInputForGrpc(a models.SuiCallArg) interface{} {
	if a == nil {
		return nil
	}
	// Pure: {type, value, valueType} -> {bytes: base64}
	if t, _ := a["type"].(string); t == "pure" {
		bytesB64 := pureValueToBytes(a)
		if bytesB64 != "" {
			return map[string]interface{}{"bytes": bytesB64}
		}
	}
	// Shared: keep initialSharedVersion, mutable, objectId only
	if _, hasInit := a["initialSharedVersion"]; hasInit {
		ver := fmt.Sprintf("%v", a["initialSharedVersion"])
		mut, _ := a["mutable"].(bool)
		objId, _ := a["objectId"].(string)
		return map[string]interface{}{
			"initialSharedVersion": ver,
			"mutable":              mut,
			"objectId":             objId,
		}
	}
	// ImmutableOrOwned: objectId + version + digest
	if objId, ok := a["objectId"].(string); ok && objId != "" {
		if ver, ok := a["version"]; ok && ver != nil {
			out := map[string]interface{}{"objectId": objId, "version": fmt.Sprintf("%v", ver)}
			if dg, ok := a["digest"]; ok && dg != nil {
				out["digest"] = dg
			}
			return out
		}
	}
	return a
}

func pureValueToBytes(a models.SuiCallArg) string {
	valueType, _ := a["valueType"].(string)
	val := a["value"]
	if val == nil {
		return ""
	}
	var b []byte
	var err error
	switch valueType {
	case "bool":
		if v, ok := val.(bool); ok {
			if v {
				b = []byte{1}
			} else {
				b = []byte{0}
			}
		}
	case "u8", "u16", "u32", "u64":
		var u uint64
		switch v := val.(type) {
		case float64:
			u = uint64(v)
		case string:
			u, _ = strconv.ParseUint(v, 10, 64)
		default:
			return ""
		}
		b, err = mystenbcs.Marshal(u)
	case "u128":
		var s string
		switch v := val.(type) {
		case string:
			s = v
		case float64:
			s = strconv.FormatUint(uint64(v), 10)
		default:
			return ""
		}
		bi := new(big.Int)
		bi.SetString(s, 10)
		be := bi.FillBytes(make([]byte, 16))
		// BCS u128 is little-endian
		for i, j := 0, len(be)-1; i < j; i, j = i+1, j-1 {
			be[i], be[j] = be[j], be[i]
		}
		b = be
	case "address", "object_id":
		if s, ok := val.(string); ok && len(s) >= 2 && s[:2] == "0x" {
			// 32 bytes hex
			if len(s) == 66 {
				b, _ = decodeHex(s[2:])
			} else {
				b, _ = decodeHex(strings.TrimLeft(s[2:], "0"))
				if len(b) < 32 {
					padded := make([]byte, 32)
					copy(padded[32-len(b):], b)
					b = padded
				}
			}
		}
	default:
		return ""
	}
	if err != nil {
		return ""
	}
	if len(b) > 0 {
		return base64.StdEncoding.EncodeToString(b)
	}
	return ""
}

func decodeHex(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		s = "0" + s
	}
	b := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		n, _ := strconv.ParseUint(s[i:i+2], 16, 8)
		b[i/2] = byte(n)
	}
	return b, nil
}

func convertObjectChangeOwner(owner interface{}) *types.ObjectOwner {
	if owner == nil {
		return nil
	}
	// Owner may be AddressOwner, ObjectOwner, Shared, etc.
	if m, ok := owner.(map[string]interface{}); ok {
		if addr, ok := m["AddressOwner"].(string); ok && addr != "" {
			return &types.ObjectOwner{Kind: types.ObjectOwnerKindAddress, AddressOwner: addr}
		}
		if addr, ok := m["ObjectOwner"].(string); ok && addr != "" {
			return &types.ObjectOwner{Kind: types.ObjectOwnerKindObject, ObjectOwner: addr}
		}
		if shared, ok := m["Shared"].(map[string]interface{}); ok {
			var ver string
			if v, ok := shared["initialSharedVersion"].(string); ok {
				ver = v
			} else if v, ok := shared["initial_shared_version"].(string); ok {
				ver = v
			} else if v, ok := shared["initialSharedVersion"].(float64); ok {
				ver = strconv.FormatFloat(v, 'f', 0, 64)
			} else if v, ok := shared["initial_shared_version"].(float64); ok {
				ver = strconv.FormatFloat(v, 'f', 0, 64)
			}
			if ver != "" {
				return &types.ObjectOwner{
					Kind:   types.ObjectOwnerKindShared,
					Shared: &types.SharedOwnerPayload{InitialSharedVersion: ver},
				}
			}
		}
	}
	return nil
}
