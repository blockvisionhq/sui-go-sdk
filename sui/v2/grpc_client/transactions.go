// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package grpc_client

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	v2proto "github.com/block-vision/sui-go-sdk/pb/sui/rpc/v2"
	. "github.com/block-vision/sui-go-sdk/sui/v2/types"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
)

// GetTransaction fetches a transaction by digest.
func (c *Client) GetTransaction(ctx context.Context, options GetTransactionOptions) (*TransactionResult, error) {
	ledgerService, err := c.grpcClient.LedgerService(ctx)
	if err != nil {
		return nil, wrapTransportError("GetTransaction", "LedgerService", err)
	}

	resp, err := ledgerService.GetTransaction(ctx, &v2proto.GetTransactionRequest{
		Digest:   &options.Digest,
		ReadMask: &fieldmaskpb.FieldMask{Paths: buildTxReadMaskPaths(options.Include)},
	})
	if err != nil {
		return nil, wrapTransportError("GetTransaction", "GetTransaction", err)
	}

	if resp.Transaction == nil {
		return nil, wrapKnownError(ErrTransactionNotFound, "GetTransaction", "missing transaction "+options.Digest, nil)
	}

	return convertGrpcExecutedTxToResult(resp.Transaction, options.Include), nil
}

// ExecuteTransaction submits a signed transaction for execution.
func (c *Client) ExecuteTransaction(ctx context.Context, options ExecuteTransactionOptions) (*TransactionResult, error) {
	execService, err := c.grpcClient.TransactionExecutionService(ctx)
	if err != nil {
		return nil, wrapTransportError("ExecuteTransaction", "TransactionExecutionService", err)
	}

	signatures := make([]*v2proto.UserSignature, len(options.Signatures))
	for i, sig := range options.Signatures {
		sigBytes, err := base64.StdEncoding.DecodeString(sig)
		if err != nil {
			return nil, fmt.Errorf("invalid signature[%d]: %w", i, err)
		}
		signatures[i] = &v2proto.UserSignature{
			Bcs: &v2proto.Bcs{Value: sigBytes},
		}
	}

	resp, err := execService.ExecuteTransaction(ctx, &v2proto.ExecuteTransactionRequest{
		Transaction: &v2proto.Transaction{
			Bcs: &v2proto.Bcs{Value: options.Transaction},
		},
		Signatures: signatures,
		ReadMask:   &fieldmaskpb.FieldMask{Paths: buildTxReadMaskPaths(options.Include)},
	})
	if err != nil {
		return nil, wrapTransportError("ExecuteTransaction", "ExecuteTransaction", err)
	}

	if resp.Transaction == nil {
		return nil, wrapKnownError(ErrTransactionNotFound, "ExecuteTransaction", "missing transaction response", nil)
	}

	return convertGrpcExecutedTxToResult(resp.Transaction, options.Include), nil
}

// SimulateTransaction simulates a transaction without submitting it.
func (c *Client) SimulateTransaction(ctx context.Context, options SimulateTransactionOptions) (*SimulateTransactionResult, error) {
	execService, err := c.grpcClient.TransactionExecutionService(ctx)
	if err != nil {
		return nil, wrapTransportError("SimulateTransaction", "TransactionExecutionService", err)
	}

	doGasSelection := false
	resp, err := execService.SimulateTransaction(ctx, &v2proto.SimulateTransactionRequest{
		Transaction:    &v2proto.Transaction{Bcs: &v2proto.Bcs{Value: options.Transaction}},
		DoGasSelection: &doGasSelection,
		ReadMask:       &fieldmaskpb.FieldMask{Paths: buildSimulateTxReadMaskPaths(options.Include)},
	})
	if err != nil {
		return nil, wrapTransportError("SimulateTransaction", "SimulateTransaction", err)
	}

	if resp.Transaction == nil {
		return nil, wrapKnownError(ErrTransactionNotFound, "SimulateTransaction", "missing transaction response", nil)
	}

	txResult := convertGrpcExecutedTxToResult(resp.Transaction, options.Include.TransactionInclude)
	result := &SimulateTransactionResult{
		Transaction: txResult.Transaction,
	}

	// Command outputs are available only when SimulateTransactionInclude.CommandResults is true.
	// TODO: convert resp.CommandOutputs when the proto API exposes them
	_ = options.Include.CommandResults

	return result, nil
}

// ─── read mask builders ────────────────────────────────────────────────────────

func buildTxReadMaskPaths(include TransactionInclude) []string {
	paths := []string{"digest", "signatures", "effects.status", "effects.epoch", "checkpoint", "timestamp"}

	if include.Transaction {
		paths = append(paths,
			"transaction.sender",
			"transaction.gas_payment",
			"transaction.expiration",
			"transaction.kind",
		)
	}
	if include.Bcs {
		paths = append(paths, "transaction.bcs")
	}
	if include.BalanceChanges {
		paths = append(paths, "balance_changes")
	}
	if include.Effects {
		paths = append(paths, "effects")
	}
	if include.Events {
		paths = append(paths, "events")
	}
	if include.ObjectTypes {
		paths = append(paths,
			"effects.changed_objects.object_type",
			"effects.changed_objects.object_id",
		)
	}
	return paths
}

func buildSimulateTxReadMaskPaths(include SimulateTransactionInclude) []string {
	base := buildTxReadMaskPaths(include.TransactionInclude)
	if include.CommandResults {
		base = append(base, "command_outputs")
	}
	return base
}

// ─── gRPC → v2 type converters ────────────────────────────────────────────────

// convertGrpcExecutedTxToResult converts a gRPC ExecutedTransaction to TransactionResult.
func convertGrpcExecutedTxToResult(tx *v2proto.ExecutedTransaction, include TransactionInclude) *TransactionResult {
	t := convertGrpcExecutedTxToTransaction(tx, include)

	result := &TransactionResult{Transaction: t}
	return result
}

// convertGrpcExecutedTxToTransaction converts the core fields.
func convertGrpcExecutedTxToTransaction(tx *v2proto.ExecutedTransaction, include TransactionInclude) *Transaction {
	t := &Transaction{
		Digest:     tx.GetDigest(),
		Checkpoint: tx.GetCheckpoint(),
	}
	if ts := tx.GetTimestamp(); ts != nil {
		t.Timestamp = ts.AsTime().UnixMilli()
	}

	// Signatures (base64)
	for _, sig := range tx.GetSignatures() {
		if bcs := sig.GetBcs(); bcs != nil && len(bcs.Value) > 0 {
			t.Signatures = append(t.Signatures, base64.StdEncoding.EncodeToString(bcs.Value))
		}
	}

	effects := tx.GetEffects()
	if effects != nil {
		status := effects.GetStatus()
		if status != nil && status.GetSuccess() {
			t.Status = ExecutionStatus{Success: true}
		} else {
			t.Status = ExecutionStatus{
				Success: false,
				Error:   convertGrpcExecutionError(status.GetError()),
			}
		}

		t.Epoch = effects.GetEpoch()

		if include.Effects {
			txEffects := convertGrpcEffects(effects)
			t.Effects = &txEffects
		}

		if include.ObjectTypes {
			objectTypes := make(map[string]string)
			for _, co := range effects.GetChangedObjects() {
				if id := co.GetObjectId(); id != "" {
					objectTypes[id] = co.GetObjectType()
				}
			}
			t.ObjectTypes = objectTypes
		}
	}

	if txProto := tx.GetTransaction(); txProto != nil {
		if include.Bcs {
			if bcs := txProto.GetBcs(); bcs != nil {
				t.BCS = bcs.GetValue()
			}
		}
		if include.Transaction {
			t.TransactionData = convertGrpcTransactionData(txProto)
		}
	}

	if include.Events {
		if evts := tx.GetEvents(); evts != nil {
			t.Events = convertGrpcEvents(evts.GetEvents())
		}
	}

	if include.BalanceChanges {
		t.BalanceChanges = convertGrpcBalanceChanges(tx.GetBalanceChanges())
	}

	return t
}

// convertGrpcTransactionData converts gRPC Transaction to type-safe TransactionData.
func convertGrpcTransactionData(tx *v2proto.Transaction) *TransactionData {
	out := &TransactionData{Sender: tx.GetSender()}

	// kind
	if k := tx.GetKind(); k != nil {
		out.Kind = k.GetKind().String()
	}

	// gasData
	if gp := tx.GetGasPayment(); gp != nil {
		payment := make([]ObjectRef, 0, len(gp.GetObjects()))
		for _, obj := range gp.GetObjects() {
			payment = append(payment, ObjectRef{
				ObjectId: obj.GetObjectId(),
				Version:  fmt.Sprintf("%d", obj.GetVersion()),
				Digest:   obj.GetDigest(),
			})
		}
		out.GasData = &GasData{
			Budget:  fmt.Sprintf("%d", gp.GetBudget()),
			Owner:   gp.GetOwner(),
			Price:   fmt.Sprintf("%d", gp.GetPrice()),
			Payment: payment,
		}
	}

	// expiration
	if exp := tx.GetExpiration(); exp != nil {
		switch exp.GetKind() {
		case v2proto.TransactionExpiration_NONE:
			out.Expiration = NoExpiration{}
		case v2proto.TransactionExpiration_EPOCH:
			out.Expiration = &EpochExpiration{Epoch: fmt.Sprintf("%d", exp.GetEpoch())}
		default:
			out.Expiration = NoExpiration{}
		}
	}

	// commands, inputs, version (from ProgrammableTransaction)
	if kind := tx.GetKind(); kind != nil {
		if pt := kind.GetProgrammableTransaction(); pt != nil {
			for _, cmd := range pt.GetCommands() {
				command := make(map[string]Command)
				c := convertGrpcCommand(cmd)
				if c != nil {
					command[commandTypeKey(cmd)] = c
				}
				out.Commands = append(out.Commands, command)
			}

			inputs := make([]Input, 0, len(pt.GetInputs()))
			for _, in := range pt.GetInputs() {
				if i := convertGrpcInput(in); i != nil {
					inputs = append(inputs, i)
				}
			}
			out.Inputs = inputs
		}
	}

	if v := tx.GetVersion(); v != 0 {
		out.Version = int(v)
	}

	return out
}

// commandKeyNames: raw proto command type names used as map keys
const (
	cmdKeyMoveCall        = "MoveCall"
	cmdKeyTransferObjects = "TransferObjects"
	cmdKeySplitCoins      = "SplitCoins"
	cmdKeyMergeCoins      = "MergeCoins"
	cmdKeyPublish         = "Publish"
	cmdKeyMakeMoveVector  = "MakeMoveVector"
	cmdKeyUpgrade         = "Upgrade"
)

func commandTypeKey(cmd *v2proto.Command) string {
	switch {
	case cmd.GetMoveCall() != nil:
		return cmdKeyMoveCall
	case cmd.GetTransferObjects() != nil:
		return cmdKeyTransferObjects
	case cmd.GetSplitCoins() != nil:
		return cmdKeySplitCoins
	case cmd.GetMergeCoins() != nil:
		return cmdKeyMergeCoins
	case cmd.GetPublish() != nil:
		return cmdKeyPublish
	case cmd.GetMakeMoveVector() != nil:
		return cmdKeyMakeMoveVector
	case cmd.GetUpgrade() != nil:
		return cmdKeyUpgrade
	default:
		return "Unknown"
	}
}

func convertGrpcCommand(cmd *v2proto.Command) Command {
	if cmd == nil {
		return nil
	}
	if mc := cmd.GetMoveCall(); mc != nil {
		return &MoveCallCommand{
			Package:       mc.GetPackage(),
			Module:        mc.GetModule(),
			Function:      mc.GetFunction(),
			TypeArguments: mc.GetTypeArguments(),
			Arguments:     convertGrpcArguments(mc.GetArguments()),
		}
	}
	if to := cmd.GetTransferObjects(); to != nil {
		return &TransferObjectsCmd{
			Objects: convertGrpcArguments(to.GetObjects()),
			Address: convertGrpcArgument(to.GetAddress()),
		}
	}
	if sc := cmd.GetSplitCoins(); sc != nil {
		return &SplitCoinsCmd{
			Coin:    convertGrpcArgument(sc.GetCoin()),
			Amounts: convertGrpcArguments(sc.GetAmounts()),
		}
	}
	if mc := cmd.GetMergeCoins(); mc != nil {
		return &MergeCoinsCmd{
			Coin:         convertGrpcArgument(mc.GetCoin()),
			CoinsToMerge: convertGrpcArguments(mc.GetCoinsToMerge()),
		}
	}
	if p := cmd.GetPublish(); p != nil {
		modules := make([]string, 0, len(p.GetModules()))
		for _, m := range p.GetModules() {
			modules = append(modules, base64.StdEncoding.EncodeToString(m))
		}
		return &PublishCmd{
			Modules:      modules,
			Dependencies: p.GetDependencies(),
		}
	}
	if mmv := cmd.GetMakeMoveVector(); mmv != nil {
		c := &MakeMoveVectorCmd{
			Elements: convertGrpcArguments(mmv.GetElements()),
		}
		if t := mmv.GetElementType(); t != "" {
			c.ElementType = t
		}
		return c
	}
	if u := cmd.GetUpgrade(); u != nil {
		modules := make([]string, 0, len(u.GetModules()))
		for _, m := range u.GetModules() {
			modules = append(modules, base64.StdEncoding.EncodeToString(m))
		}
		c := &UpgradeCmd{
			Modules:      modules,
			Dependencies: u.GetDependencies(),
			Package:      u.GetPackage(),
		}
		if t := convertGrpcArgument(u.GetTicket()); t != nil {
			c.UpgradeTicket = t
		}
		return c
	}
	return nil
}

func convertGrpcArguments(args []*v2proto.Argument) []Argument {
	out := make([]Argument, 0, len(args))
	for _, a := range args {
		if arg := convertGrpcArgument(a); arg != nil {
			out = append(out, arg)
		}
	}
	return out
}

func convertGrpcArgument(a *v2proto.Argument) Argument {
	if a == nil {
		return nil
	}
	switch a.GetKind() {
	case v2proto.Argument_GAS:
		return GasCoinArg{}
	case v2proto.Argument_INPUT:
		return &InputArg{Index: int(a.GetInput())}
	case v2proto.Argument_RESULT:
		arg := &ResultArg{Result: int(a.GetResult())}
		if s := a.GetSubresult(); s != 0 {
			ss := int(s)
			arg.Subresult = &ss
		}
		return arg
	default:
		return nil
	}
}

func convertGrpcInput(in *v2proto.Input) Input {
	if in == nil {
		return nil
	}
	// literal and pure are mutually exclusive (oneof); prefer literal for typed representation
	if lit := in.GetLiteral(); lit != nil {
		if p := convertLiteralToPureInput(lit); p != nil {
			return p
		}
	}
	switch in.GetKind() {
	case v2proto.Input_PURE:
		bytes := ""
		if b := in.GetPure(); len(b) > 0 {
			bytes = base64.StdEncoding.EncodeToString(b)
		}
		return &PureBCSInput{Bytes: bytes}
	case v2proto.Input_SHARED:
		return &SharedObjectInput{
			ObjectId:             in.GetObjectId(),
			InitialSharedVersion: strconv.FormatUint(in.GetVersion(), 10),
			Mutable:              in.GetMutable(),
		}
	case v2proto.Input_IMMUTABLE_OR_OWNED:
		return &ImmutableOrOwnedInput{
			ObjectId: in.GetObjectId(),
			Version:  strconv.FormatUint(in.GetVersion(), 10),
			Digest:   in.GetDigest(),
		}
	case v2proto.Input_RECEIVING:
		return &ReceivingInput{
			ObjectId: in.GetObjectId(),
			Version:  strconv.FormatUint(in.GetVersion(), 10),
			Digest:   in.GetDigest(),
		}
	default:
		return nil
	}
}

// convertGrpcEffects converts gRPC TransactionEffects to the v2 TransactionEffects type.
func convertGrpcEffects(effects *v2proto.TransactionEffects) TransactionEffects {
	e := TransactionEffects{
		TransactionDigest: effects.GetTransactionDigest(),
		Dependencies:      effects.GetDependencies(),
	}

	status := effects.GetStatus()
	if status != nil {
		if status.GetSuccess() {
			e.Status = ExecutionStatus{Success: true}
		} else {
			e.Status = ExecutionStatus{
				Success: false,
				Error:   convertGrpcExecutionError(status.GetError()),
			}
		}
	}

	if gasUsed := effects.GetGasUsed(); gasUsed != nil {
		e.GasUsed = GasCostSummary{
			ComputationCost:         fmt.Sprintf("%d", gasUsed.GetComputationCost()),
			StorageCost:             fmt.Sprintf("%d", gasUsed.GetStorageCost()),
			StorageRebate:           fmt.Sprintf("%d", gasUsed.GetStorageRebate()),
			NonRefundableStorageFee: fmt.Sprintf("%d", gasUsed.GetNonRefundableStorageFee()),
		}
	}

	if ed := effects.GetEventsDigest(); ed != "" {
		e.EventsDigest = &ed
	}

	for _, co := range effects.GetChangedObjects() {
		e.ChangedObjects = append(e.ChangedObjects, convertGrpcChangedObject(co))
	}

	if gas := effects.GetGasObject(); gas != nil {
		co := convertGrpcChangedObject(gas)
		e.GasObject = &co
	}

	e.ExecutedEpoch = effects.GetEpoch()

	return e
}

// convertGrpcChangedObject converts a gRPC ChangedObject to the v2 ChangedObject type.
func convertGrpcChangedObject(co *v2proto.ChangedObject) ChangedObject {
	obj := ChangedObject{
		ObjectId:    co.GetObjectId(),
		InputState:  grpcInputState(co.GetInputState()),
		OutputState: grpcOutputState(co.GetOutputState()),
		IdOperation: grpcIdOperation(co.GetIdOperation()),
	}

	if v := co.GetInputVersion(); v != 0 {
		s := fmt.Sprintf("%d", v)
		obj.InputVersion = &s
	}
	if d := co.GetInputDigest(); d != "" {
		obj.InputDigest = &d
	}
	if owner := co.GetInputOwner(); owner != nil {
		o := convertGrpcOwner(owner)
		obj.InputOwner = &o
	}
	if v := co.GetOutputVersion(); v != 0 {
		s := fmt.Sprintf("%d", v)
		obj.OutputVersion = &s
	}
	if d := co.GetOutputDigest(); d != "" {
		obj.OutputDigest = &d
	}
	if owner := co.GetOutputOwner(); owner != nil {
		o := convertGrpcOwner(owner)
		obj.OutputOwner = &o
	}

	return obj
}

func grpcInputState(s v2proto.ChangedObject_InputObjectState) ChangedObjectInputState {
	switch s {
	case v2proto.ChangedObject_INPUT_OBJECT_STATE_DOES_NOT_EXIST:
		return ChangedObjectInputDoesNotExist
	case v2proto.ChangedObject_INPUT_OBJECT_STATE_EXISTS:
		return ChangedObjectInputExists
	default:
		return ChangedObjectInputUnknown
	}
}

func grpcOutputState(s v2proto.ChangedObject_OutputObjectState) ChangedObjectOutputState {
	switch s {
	case v2proto.ChangedObject_OUTPUT_OBJECT_STATE_DOES_NOT_EXIST:
		return ChangedObjectOutputDoesNotExist
	case v2proto.ChangedObject_OUTPUT_OBJECT_STATE_OBJECT_WRITE:
		return ChangedObjectOutputObjectWrite
	case v2proto.ChangedObject_OUTPUT_OBJECT_STATE_PACKAGE_WRITE:
		return ChangedObjectOutputPackageWrite
	default:
		return ChangedObjectOutputUnknown
	}
}

func grpcIdOperation(op v2proto.ChangedObject_IdOperation) ChangedObjectIdOperation {
	switch op {
	case v2proto.ChangedObject_NONE:
		return ChangedObjectIdOpNone
	case v2proto.ChangedObject_CREATED:
		return ChangedObjectIdOpCreated
	case v2proto.ChangedObject_DELETED:
		return ChangedObjectIdOpDeleted
	default:
		return ChangedObjectIdOpUnknown
	}
}

// convertGrpcExecutionError converts a gRPC ExecutionError to the v2 ExecutionError type.
func convertGrpcExecutionError(err *v2proto.ExecutionError) *ExecutionError {
	if err == nil {
		return nil
	}
	msg := err.GetDescription()
	if msg == "" {
		msg = err.GetKind().String()
	}
	return &ExecutionError{
		Message: msg,
		Kind:    ExecutionErrorKindUnknown,
	}
}

// convertGrpcEvents converts gRPC []*Event to the v2 []Event type.
func convertGrpcEvents(events []*v2proto.Event) []Event {
	if len(events) == 0 {
		return nil
	}
	result := make([]Event, len(events))
	for i, ev := range events {
		result[i] = Event{
			PackageId: ev.GetPackageId(),
			Module:    ev.GetModule(),
			Sender:    ev.GetSender(),
			EventType: ev.GetEventType(),
			BCS:       ev.GetContents().GetValue(),
		}
		if j := ev.GetJson(); j != nil {
			result[i].JSON = structpbValueToMap(j)
		}
	}
	return result
}

// convertGrpcBalanceChanges converts gRPC []*BalanceChange to the v2 []BalanceChange type.
func convertGrpcBalanceChanges(changes []*v2proto.BalanceChange) []BalanceChange {
	if len(changes) == 0 {
		return nil
	}
	result := make([]BalanceChange, len(changes))
	for i, ch := range changes {
		result[i] = BalanceChange{
			CoinType: ch.GetCoinType(),
			Address:  ch.GetAddress(),
			Amount:   ch.GetAmount(),
		}
	}
	return result
}

// convertLiteralToPureInput converts proto literal (structpb.Value) to the corresponding Pure*Input.
// Returns nil when unmappable; caller falls back to PureBCSInput.
func convertLiteralToPureInput(v *structpb.Value) Input {
	if v == nil {
		return nil
	}
	switch v.Kind.(type) {
	case *structpb.Value_NullValue:
		return nil
	case *structpb.Value_NumberValue:
		n := v.GetNumberValue()
		u := uint64(n)
		if n >= 0 && float64(u) == n && n < 1<<64 {
			return &PureU64Input{Value: u}
		}
		return &PureLiteralInput{Value: n}
	case *structpb.Value_StringValue:
		s := v.GetStringValue()
		if len(s) >= 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') && len(s) <= 66 {
			return &PureAddressInput{Value: s}
		}
		return &PureStringInput{Value: s}
	case *structpb.Value_BoolValue:
		return &PureBoolInput{Value: v.GetBoolValue()}
	case *structpb.Value_ListValue:
		list := v.GetListValue().GetValues()
		els := make([]Input, 0, len(list))
		for _, item := range list {
			if item == nil {
				continue
			}
			if _, ok := item.Kind.(*structpb.Value_NullValue); ok {
				els = append(els, &PureOptionInput{}) // None
				continue
			}
			if p := convertLiteralToPureInput(item); p != nil {
				els = append(els, p)
			} else {
				els = append(els, &PureLiteralInput{Value: structpbValueToInterface(item)})
			}
		}
		return &PureVecInput{Elements: els}
	case *structpb.Value_StructValue:
		return &PureLiteralInput{Value: structpbValueToInterface(v)}
	default:
		return nil
	}
}

// ─── structpb helpers (shared by objects.go and transactions.go) ──────────────

func structpbValueToMap(v *structpb.Value) map[string]interface{} {
	if v == nil {
		return nil
	}
	sv := v.GetStructValue()
	if sv == nil {
		return nil
	}
	result := make(map[string]interface{}, len(sv.GetFields()))
	for k, val := range sv.GetFields() {
		result[k] = structpbValueToInterface(val)
	}
	return result
}

func structpbValueToInterface(v *structpb.Value) interface{} {
	if v == nil {
		return nil
	}
	switch v.Kind.(type) {
	case *structpb.Value_NullValue:
		return nil
	case *structpb.Value_NumberValue:
		return v.GetNumberValue()
	case *structpb.Value_StringValue:
		return v.GetStringValue()
	case *structpb.Value_BoolValue:
		return v.GetBoolValue()
	case *structpb.Value_StructValue:
		return structpbValueToMap(v)
	case *structpb.Value_ListValue:
		list := v.GetListValue().GetValues()
		out := make([]interface{}, len(list))
		for i, item := range list {
			out[i] = structpbValueToInterface(item)
		}
		return out
	default:
		return nil
	}
}
