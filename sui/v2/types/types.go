// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package types

import "encoding/json"

// SUI_TYPE_ARG is the type argument for native SUI coin.
const SUI_TYPE_ARG = "0x2::sui::SUI"

// ═══════════════════════════════════════════════════════════════════════════════
// ObjectOwner
// ═══════════════════════════════════════════════════════════════════════════════

// ObjectOwnerKind identifies which variant of ObjectOwner is active.
type ObjectOwnerKind string

const (
	ObjectOwnerKindAddress          ObjectOwnerKind = "AddressOwner"
	ObjectOwnerKindObject           ObjectOwnerKind = "ObjectOwner"
	ObjectOwnerKindShared           ObjectOwnerKind = "Shared"
	ObjectOwnerKindImmutable        ObjectOwnerKind = "Immutable"
	ObjectOwnerKindConsensusAddress ObjectOwnerKind = "ConsensusAddressOwner"
	ObjectOwnerKindUnknown          ObjectOwnerKind = "Unknown"
)

// ObjectOwner is a discriminated union that describes who owns a Sui object.
// Inspect Kind to determine which fields are populated:
//
//   - ObjectOwnerKindAddress    → AddressOwner
//   - ObjectOwnerKindObject     → ObjectOwner
//   - ObjectOwnerKindShared     → Shared
//   - ObjectOwnerKindImmutable  → (no payload)
//   - ObjectOwnerKindConsensusAddress → ConsensusAddress
//   - ObjectOwnerKindUnknown    → (no payload)
type ObjectOwner struct {
	Kind             ObjectOwnerKind
	AddressOwner     string                 // Kind == ObjectOwnerKindAddress
	ObjectOwner      string                 // Kind == ObjectOwnerKindObject
	Shared           *SharedOwnerPayload    // Kind == ObjectOwnerKindShared
	ConsensusAddress *ConsensusOwnerPayload // Kind == ObjectOwnerKindConsensusAddress
}

func (o ObjectOwner) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{"kind": string(o.Kind)}
	switch o.Kind {
	case ObjectOwnerKindAddress:
		m["AddressOwner"] = o.AddressOwner
	case ObjectOwnerKindObject:
		m["ObjectOwner"] = o.ObjectOwner
	case ObjectOwnerKindShared:
		if o.Shared != nil {
			m["Shared"] = o.Shared
		}
	case ObjectOwnerKindImmutable:
	case ObjectOwnerKindConsensusAddress:
		if o.ConsensusAddress != nil {
			m["ConsensusAddressOwner"] = o.ConsensusAddress
		}
	case ObjectOwnerKindUnknown:
	}
	return json.Marshal(m)
}

// SharedOwnerPayload holds the initial shared version for a shared object.
type SharedOwnerPayload struct {
	InitialSharedVersion string `json:"initialSharedVersion"`
}

// ConsensusOwnerPayload holds consensus-address-owner details.
type ConsensusOwnerPayload struct {
	StartVersion string `json:"startVersion"`
	Owner        string `json:"owner"`
}

// ObjectInclude controls which optional fields are returned in an object
// response.  All fields default to false (not requested).
type ObjectInclude struct {
	// Include the BCS-encoded Move struct content.
	// Parse with generated BCS type parsers (not bcs.Object).
	Content bool `json:"content,omitempty"`

	// Include the digest of the transaction that last mutated the object.
	PreviousTransaction bool `json:"previousTransaction,omitempty"`

	// Include the full BCS-encoded object envelope.
	// Parse with bcs.Object, not a Move struct parser.
	ObjectBcs bool `json:"objectBcs,omitempty"`

	// Include the JSON representation of the Move struct content.
	// Warning: shape may differ between JSON-RPC and gRPC implementations.
	JSON bool `json:"json,omitempty"`
}

// Object is a Sui object with optional fields controlled by ObjectInclude.
type Object struct {
	ObjectId string      `json:"objectId"`
	Version  string      `json:"version"`
	Digest   string      `json:"digest"`
	Owner    ObjectOwner `json:"owner"`
	Type     string      `json:"type"`

	// Populated only when ObjectInclude.Content == true.
	// Raw BCS bytes of the object's Move struct — pass to generated BCS parsers.
	Content []byte `json:"content,omitempty"`

	// Populated only when ObjectInclude.PreviousTransaction == true.
	PreviousTransaction *string `json:"previousTransaction,omitempty"`

	// Populated only when ObjectInclude.ObjectBcs == true.
	// Full BCS-encoded object envelope — parse with bcs.Object, not a struct parser.
	ObjectBcs []byte `json:"objectBcs,omitempty"`

	// Populated only when ObjectInclude.JSON == true.
	// Warning: exact shape may vary between API implementations.
	JSON map[string]interface{} `json:"json,omitempty"`
}

// Coin is a simplified object view for Coin<T> objects.
type Coin struct {
	ObjectId string      `json:"objectId"`
	Version  string      `json:"version"`
	Digest   string      `json:"digest"`
	Owner    ObjectOwner `json:"owner"`
	Type     string      `json:"type"`
	Balance  string      `json:"balance"`
}

// GetObjectsOptions is the request for fetching multiple objects by ID.
type GetObjectsOptions struct {
	ObjectIds []string      `json:"objectIds"`
	Include   ObjectInclude `json:"include,omitempty"`
}

// GetObjectsResponse is the response for GetObjects.
// Error entries represent objects that could not be retrieved.
type GetObjectsResponse struct {
	Objects []ObjectOrError `json:"objects"`
}

// ObjectOrError holds either a successfully retrieved object or an error
// message for that object ID.
type ObjectOrError struct {
	Object *Object `json:"object,omitempty"`
	Error  *string `json:"error,omitempty"`
}

// ListOwnedObjectsOptions is the request for listing objects owned by an address.
type ListOwnedObjectsOptions struct {
	Owner   string        `json:"owner"`
	Limit   *uint32       `json:"limit,omitempty"`
	Cursor  *string       `json:"cursor,omitempty"`
	Type    *string       `json:"type,omitempty"`
	Include ObjectInclude `json:"include,omitempty"`
}

// ListOwnedObjectsResponse is the response for ListOwnedObjects.
type ListOwnedObjectsResponse struct {
	Objects     []Object `json:"objects"`
	HasNextPage bool     `json:"hasNextPage"`
	Cursor      *string  `json:"cursor"`
}

// ListCoinsOptions is the request for listing Coin<T> objects.
type ListCoinsOptions struct {
	Owner    string  `json:"owner"`
	CoinType *string `json:"coinType,omitempty"` // defaults to 0x2::sui::SUI
	Limit    *uint32 `json:"limit,omitempty"`
	Cursor   *string `json:"cursor,omitempty"`
}

// ListCoinsResponse is the response for ListCoins.
type ListCoinsResponse struct {
	Objects     []Coin  `json:"objects"`
	HasNextPage bool    `json:"hasNextPage"`
	Cursor      *string `json:"cursor"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// Dynamic fields
// ═══════════════════════════════════════════════════════════════════════════════

// DynamicFieldName identifies a dynamic field by its Move type and BCS-encoded
// name value.
type DynamicFieldName struct {
	Type string `json:"type"`
	BCS  []byte `json:"bcs"`
}

// DynamicFieldValue holds the type and BCS bytes of a dynamic field's value.
type DynamicFieldValue struct {
	Type string `json:"type"`
	BCS  []byte `json:"bcs"`
}

// DynamicFieldKind distinguishes plain dynamic fields from dynamic object fields.
type DynamicFieldKind string

const (
	DynamicFieldKindDynamicField  DynamicFieldKind = "DynamicField"
	DynamicFieldKindDynamicObject DynamicFieldKind = "DynamicObject"
)

// DynamicFieldEntry is the lightweight descriptor returned by ListDynamicFields.
type DynamicFieldEntry struct {
	FieldId   string           `json:"fieldId"`
	Type      string           `json:"type"`
	Name      DynamicFieldName `json:"name"`
	ValueType string           `json:"valueType"`
	// Kind indicates whether this is a DynamicField or a DynamicObject.
	Kind        DynamicFieldKind `json:"kind"`
	FieldObject Object           `json:"fieldObject,omitempty"`
	ChildId     *string          `json:"childId,omitempty"` // non-nil for DynamicObject
}

// DynamicField is the full dynamic field response including its value.
type DynamicField struct {
	DynamicFieldEntry
	Value               DynamicFieldValue `json:"value"`
	Version             string            `json:"version"`
	Digest              string            `json:"digest"`
	PreviousTransaction *string           `json:"previousTransaction"`
}

// ListDynamicFieldsOptions is the request for listing dynamic fields.
type ListDynamicFieldsOptions struct {
	ParentId string  `json:"parentId"`
	Limit    *uint32 `json:"limit,omitempty"`
	Cursor   *string `json:"cursor,omitempty"`
}

// ListDynamicFieldsResponse is the response for ListDynamicFields.
type ListDynamicFieldsResponse struct {
	HasNextPage   bool                `json:"hasNextPage"`
	Cursor        *string             `json:"cursor"`
	DynamicFields []DynamicFieldEntry `json:"dynamicFields"`
}

// GetDynamicFieldOptions is the request for fetching a specific dynamic field.
type GetDynamicFieldOptions struct {
	ParentId string           `json:"parentId"`
	Name     DynamicFieldName `json:"name"`
}

// GetDynamicFieldResponse is the response for GetDynamicField.
type GetDynamicFieldResponse struct {
	DynamicField DynamicField `json:"dynamicField"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// Balance
// ═══════════════════════════════════════════════════════════════════════════════

// Balance represents the aggregated balance of one coin type for an address.
type Balance struct {
	CoinType string `json:"coinType"`
	// Total balance across all Coin<T> objects owned by the address.
	Balance_ string `json:"balance"`
	// Balance held in Coin<T> objects (same as Balance for most coins).
	CoinBalance string `json:"coinBalance"`
	// Balance held directly at the address level (e.g., SUI gas balance).
	AddressBalance string `json:"addressBalance"`
}

// GetBalanceOptions is the request for GetBalance.
type GetBalanceOptions struct {
	Owner    string  `json:"owner"`
	CoinType *string `json:"coinType,omitempty"` // defaults to 0x2::sui::SUI
}

// GetBalanceResponse is the response for GetBalance.
type GetBalanceResponse struct {
	Balance Balance `json:"balance"`
}

// ListBalancesOptions is the request for ListBalances.
type ListBalancesOptions struct {
	Owner  string  `json:"owner"`
	Limit  *uint32 `json:"limit,omitempty"`
	Cursor *string `json:"cursor,omitempty"`
}

// ListBalancesResponse is the response for ListBalances.
type ListBalancesResponse struct {
	Balances    []Balance `json:"balances"`
	HasNextPage bool      `json:"hasNextPage"`
	Cursor      *string   `json:"cursor"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// Coin metadata
// ═══════════════════════════════════════════════════════════════════════════════

// CoinMetadata holds on-chain metadata for a fungible token type.
type CoinMetadata struct {
	// Object ID of the CoinMetadata object on chain; nil for native SUI.
	Id          *string `json:"id"`
	Decimals    int     `json:"decimals"`
	Name        string  `json:"name"`
	Symbol      string  `json:"symbol"`
	Description string  `json:"description"`
	IconUrl     *string `json:"iconUrl"`
}

// GetCoinMetadataOptions is the request for GetCoinMetadata.
type GetCoinMetadataOptions struct {
	CoinType string `json:"coinType"`
}

// GetCoinMetadataResponse is the response for GetCoinMetadata.
type GetCoinMetadataResponse struct {
	// nil when the coin type has no on-chain metadata object.
	CoinMetadata *CoinMetadata `json:"coinMetadata"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// Execution status & effects
// ═══════════════════════════════════════════════════════════════════════════════

// ExecutionStatus reports whether a transaction succeeded or failed.
type ExecutionStatus struct {
	Success bool            `json:"success"`
	Error   *ExecutionError `json:"error,omitempty"` // non-nil when Success == false
}

// ExecutionErrorKind enumerates the top-level categories of execution failure.
type ExecutionErrorKind string

const (
	ExecutionErrorKindMoveAbort            ExecutionErrorKind = "MoveAbort"
	ExecutionErrorKindSizeError            ExecutionErrorKind = "SizeError"
	ExecutionErrorKindCommandArgumentError ExecutionErrorKind = "CommandArgumentError"
	ExecutionErrorKindTypeArgumentError    ExecutionErrorKind = "TypeArgumentError"
	ExecutionErrorKindPackageUpgradeError  ExecutionErrorKind = "PackageUpgradeError"
	ExecutionErrorKindIndexError           ExecutionErrorKind = "IndexError"
	ExecutionErrorKindCoinDenyListError    ExecutionErrorKind = "CoinDenyListError"
	ExecutionErrorKindCongestedObjects     ExecutionErrorKind = "CongestedObjects"
	ExecutionErrorKindObjectIdError        ExecutionErrorKind = "ObjectIdError"
	ExecutionErrorKindUnknown              ExecutionErrorKind = "Unknown"
)

// ExecutionError carries the error kind and detail for a failed execution.
// Inspect Kind to determine which payload field is populated.
type ExecutionError struct {
	Message string             `json:"message"`
	Command *uint32            `json:"command,omitempty"`
	Kind    ExecutionErrorKind `json:"kind"`

	MoveAbort            *MoveAbort            `json:"MoveAbort,omitempty"`
	SizeError            *SizeError            `json:"SizeError,omitempty"`
	CommandArgumentError *CommandArgumentError `json:"CommandArgumentError,omitempty"`
	TypeArgumentError    *TypeArgumentError    `json:"TypeArgumentError,omitempty"`
	PackageUpgradeError  *PackageUpgradeError  `json:"PackageUpgradeError,omitempty"`
	IndexError           *IndexError           `json:"IndexError,omitempty"`
	CoinDenyListError    *CoinDenyListError    `json:"CoinDenyListError,omitempty"`
	CongestedObjects     *CongestedObjects     `json:"CongestedObjects,omitempty"`
	ObjectIdError        *ObjectIdError        `json:"ObjectIdError,omitempty"`
}

// MoveAbort is the payload for a Move abort execution error.
type MoveAbort struct {
	AbortCode   string        `json:"abortCode"`
	Location    *MoveLocation `json:"location,omitempty"`
	CleverError *CleverError  `json:"cleverError,omitempty"`
}

// MoveLocation pinpoints the Move instruction that aborted.
type MoveLocation struct {
	Package      string `json:"package"`
	Module       string `json:"module"`
	Function     uint32 `json:"function"`
	FunctionName string `json:"functionName"`
	Instruction  uint32 `json:"instruction"`
}

// CleverError carries structured error information from a Move abort.
type CleverError struct {
	ErrorCode    uint32 `json:"errorCode"`
	LineNumber   uint32 `json:"lineNumber"`
	ConstantName string `json:"constantName"`
	ConstantType string `json:"constantType"`
	Value        string `json:"value"`
}

// SizeError is the payload for an object-size limit violation.
type SizeError struct {
	Name    string `json:"name"`
	Size    int    `json:"size"`
	MaxSize int    `json:"maxSize"`
}

// CommandArgumentError identifies a bad argument at a specific command position.
type CommandArgumentError struct {
	Argument int    `json:"argument"`
	Name     string `json:"name"`
}

// TypeArgumentError identifies a bad type argument.
type TypeArgumentError struct {
	TypeArgument int    `json:"typeArgument"`
	Name         string `json:"name"`
}

// PackageUpgradeError is the payload for a package upgrade failure.
type PackageUpgradeError struct {
	Name      string  `json:"name"`
	PackageId *string `json:"packageId,omitempty"`
	Digest    *string `json:"digest,omitempty"`
}

// IndexError is the payload for an out-of-range index access.
type IndexError struct {
	Index     *int `json:"index,omitempty"`
	Subresult *int `json:"subresult,omitempty"`
}

// CoinDenyListError is the payload for a coin deny-list violation.
type CoinDenyListError struct {
	Name     string  `json:"name"`
	CoinType string  `json:"coinType"`
	Address  *string `json:"address,omitempty"`
}

// CongestedObjects is the payload when execution fails due to object congestion.
type CongestedObjects struct {
	Name    string   `json:"name"`
	Objects []string `json:"objects"`
}

// ObjectIdError is the payload for an invalid object ID.
type ObjectIdError struct {
	Name     *string `json:"name,omitempty"`
	ObjectId string  `json:"objectId"`
}

// GasCostSummary breaks down the gas consumed by a transaction.
type GasCostSummary struct {
	ComputationCost         string `json:"computationCost"`
	StorageCost             string `json:"storageCost"`
	StorageRebate           string `json:"storageRebate"`
	NonRefundableStorageFee string `json:"nonRefundableStorageFee"`
}

// ChangedObjectInputState describes an object's state before the transaction.
type ChangedObjectInputState string

const (
	ChangedObjectInputUnknown      ChangedObjectInputState = "Unknown"
	ChangedObjectInputDoesNotExist ChangedObjectInputState = "DoesNotExist"
	ChangedObjectInputExists       ChangedObjectInputState = "Exists"
)

// ChangedObjectOutputState describes how the object changed after the transaction.
type ChangedObjectOutputState string

const (
	ChangedObjectOutputUnknown            ChangedObjectOutputState = "Unknown"
	ChangedObjectOutputDoesNotExist       ChangedObjectOutputState = "DoesNotExist"
	ChangedObjectOutputObjectWrite        ChangedObjectOutputState = "ObjectWrite"
	ChangedObjectOutputPackageWrite       ChangedObjectOutputState = "PackageWrite"
	ChangedObjectOutputAccumulatorWriteV1 ChangedObjectOutputState = "AccumulatorWriteV1"
)

// ChangedObjectIdOperation describes whether an object was created or deleted.
type ChangedObjectIdOperation string

const (
	ChangedObjectIdOpUnknown ChangedObjectIdOperation = "Unknown"
	ChangedObjectIdOpNone    ChangedObjectIdOperation = "None"
	ChangedObjectIdOpCreated ChangedObjectIdOperation = "Created"
	ChangedObjectIdOpDeleted ChangedObjectIdOperation = "Deleted"
)

// ChangedObject records the state transition of one object in transaction effects.
type ChangedObject struct {
	ObjectId      string                   `json:"objectId"`
	InputState    ChangedObjectInputState  `json:"inputState"`
	InputVersion  *string                  `json:"inputVersion"`
	InputDigest   *string                  `json:"inputDigest"`
	InputOwner    *ObjectOwner             `json:"inputOwner"`
	OutputState   ChangedObjectOutputState `json:"outputState"`
	OutputVersion *string                  `json:"outputVersion"`
	OutputDigest  *string                  `json:"outputDigest"`
	OutputOwner   *ObjectOwner             `json:"outputOwner"`
	IdOperation   ChangedObjectIdOperation `json:"idOperation"`
}

// UnchangedConsensusObjectKind describes why a consensus object was not mutated.
type UnchangedConsensusObjectKind string

const (
	UnchangedConsensusObjectKindUnknown                    UnchangedConsensusObjectKind = "Unknown"
	UnchangedConsensusObjectKindReadOnlyRoot               UnchangedConsensusObjectKind = "ReadOnlyRoot"
	UnchangedConsensusObjectKindMutateConsensusStreamEnded UnchangedConsensusObjectKind = "MutateConsensusStreamEnded"
	UnchangedConsensusObjectKindReadConsensusStreamEnded   UnchangedConsensusObjectKind = "ReadConsensusStreamEnded"
	UnchangedConsensusObjectKindCancelled                  UnchangedConsensusObjectKind = "Cancelled"
	UnchangedConsensusObjectKindPerEpochConfig             UnchangedConsensusObjectKind = "PerEpochConfig"
)

// UnchangedConsensusObject records a consensus object that was read but not mutated.
type UnchangedConsensusObject struct {
	Kind     UnchangedConsensusObjectKind `json:"kind"`
	ObjectId string                       `json:"objectId"`
	Version  *string                      `json:"version"`
	Digest   *string                      `json:"digest"`
}

// TransactionEffects carries the full on-chain effects of an executed transaction.
type TransactionEffects struct {
	Version                   int                        `json:"version"`
	Status                    ExecutionStatus            `json:"status"`
	GasUsed                   GasCostSummary             `json:"gasUsed"`
	ExecutedEpoch             uint64                     `json:"executedEpoch"`
	TransactionDigest         string                     `json:"transactionDigest"`
	GasObject                 *ChangedObject             `json:"gasObject"`
	EventsDigest              *string                    `json:"eventsDigest"`
	Dependencies              []string                   `json:"dependencies"`
	ChangedObjects            []ChangedObject            `json:"changedObjects"`
	UnchangedConsensusObjects []UnchangedConsensusObject `json:"unchangedConsensusObjects"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// Events
// ═══════════════════════════════════════════════════════════════════════════════

// Event represents a single Move event emitted by a transaction.
type Event struct {
	PackageId string `json:"packageId"`
	Module    string `json:"module"`
	Sender    string `json:"sender"`
	EventType string `json:"eventType"`
	// Raw BCS bytes of the event's Move struct — pass to generated BCS parsers.
	BCS []byte `json:"bcs"`
	// JSON representation of the event's Move struct data.
	// Warning: exact shape may vary between API implementations.
	JSON map[string]interface{} `json:"json,omitempty"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// Balance changes
// ═══════════════════════════════════════════════════════════════════════════════

// BalanceChange records a net coin balance change caused by a transaction.
type BalanceChange struct {
	CoinType string `json:"coinType"`
	Address  string `json:"address"`
	// Signed decimal string — positive for gains, negative for losses.
	Amount string `json:"amount"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// Transaction
// ═══════════════════════════════════════════════════════════════════════════════

// TransactionInclude controls which optional fields are returned in a
// transaction response.
type TransactionInclude struct {
	// Include balance changes caused by the transaction.
	BalanceChanges bool `json:"balanceChanges,omitempty"`
	// Include parsed transaction effects.
	Effects bool `json:"effects,omitempty"`
	// Include events emitted by the transaction.
	Events bool `json:"events,omitempty"`
	// Include a map of objectId → type for all changed objects.
	ObjectTypes bool `json:"objectTypes,omitempty"`
	// Include the parsed transaction data (sender, gas, inputs, commands).
	Transaction bool `json:"transaction,omitempty"`
	// Include the raw BCS-encoded transaction bytes.
	Bcs bool `json:"bcs,omitempty"`
}

// SimulateTransactionInclude extends TransactionInclude with simulation-only
// command results.
type SimulateTransactionInclude struct {
	TransactionInclude
	// Include return values and mutated references from each command.
	CommandResults bool `json:"commandResults,omitempty"`
}

// CommandOutput holds BCS bytes for a command return value or mutated reference.
type CommandOutput struct {
	BCS []byte `json:"bcs"`
}

// CommandResult holds the return values and mutated references from one command.
type CommandResult struct {
	ReturnValues      []CommandOutput `json:"returnValues"`
	MutatedReferences []CommandOutput `json:"mutatedReferences"`
}

// TransactionData is the parsed Programmable Transaction Block data.
type TransactionData struct {
	GasData    *GasData              `json:"gasData,omitempty"`
	Sender     string                `json:"sender"`
	Expiration TransactionExpiration `json:"expiration,omitempty"`
	Kind       string                `json:"kind"`
	Commands   []map[string]Command  `json:"commands,omitempty"`
	Inputs     []Input               `json:"inputs,omitempty"`
	Version    int                   `json:"version,omitempty"`
}

// GasData holds transaction gas configuration.
type GasData struct {
	Budget  string      `json:"budget"`
	Owner   string      `json:"owner"`
	Price   string      `json:"price"`
	Payment []ObjectRef `json:"payment,omitempty"`
}

// ObjectRef is an object reference.
type ObjectRef struct {
	ObjectId string `json:"objectId"`
	Version  string `json:"version"`
	Digest   string `json:"digest"`
}

// TransactionExpiration is the transaction expiration; use type assertion to distinguish.
// Implementations: *NoExpiration, *EpochExpiration
type TransactionExpiration interface {
	isTransactionExpiration()
}

// NoExpiration means no expiration.
type NoExpiration struct{}

func (NoExpiration) isTransactionExpiration() {}

// EpochExpiration expires at a specific epoch.
type EpochExpiration struct {
	Epoch string `json:"epoch"`
}

func (*EpochExpiration) isTransactionExpiration() {}

// Command is a PTB command; use type assertion to distinguish.
// Implementations: *MoveCallCommand, *TransferObjectsCmd, *SplitCoinsCmd, *MergeCoinsCmd, *PublishCmd, *MakeMoveVectorCmd, *UpgradeCmd
type Command interface {
	isCommand()
}

// MoveCallCommand is a Move call command.
type MoveCallCommand struct {
	Package       string     `json:"package"`
	Module        string     `json:"module"`
	Function      string     `json:"function"`
	TypeArguments []string   `json:"typeArguments,omitempty"`
	Arguments     []Argument `json:"arguments,omitempty"`
}

func (*MoveCallCommand) isCommand() {}

// TransferObjectsCmd transfers objects.
type TransferObjectsCmd struct {
	Objects []Argument `json:"objects"`
	Address Argument   `json:"address,omitempty"`
}

func (*TransferObjectsCmd) isCommand() {}

// SplitCoinsCmd splits coins.
type SplitCoinsCmd struct {
	Coin    Argument   `json:"coin,omitempty"`
	Amounts []Argument `json:"amounts,omitempty"`
}

func (*SplitCoinsCmd) isCommand() {}

// MergeCoinsCmd merges coins.
type MergeCoinsCmd struct {
	Coin         Argument   `json:"coin,omitempty"`
	CoinsToMerge []Argument `json:"coinsToMerge,omitempty"`
}

func (*MergeCoinsCmd) isCommand() {}

type PublishCmd struct {
	Modules      []string `json:"modules"`
	Dependencies []string `json:"dependencies,omitempty"`
}

func (*PublishCmd) isCommand() {}

// MakeMoveVectorCmd builds a vector.
type MakeMoveVectorCmd struct {
	ElementType string     `json:"elementType,omitempty"`
	Elements    []Argument `json:"elements,omitempty"`
}

func (*MakeMoveVectorCmd) isCommand() {}

// UpgradeCmd upgrades a package.
type UpgradeCmd struct {
	Modules       []string `json:"modules"`
	Dependencies  []string `json:"dependencies,omitempty"`
	Package       string   `json:"package"`
	UpgradeTicket Argument `json:"upgradeTicket,omitempty"`
}

func (*UpgradeCmd) isCommand() {}

// RawCommand holds arbitrary command JSON for pass-through (e.g. from JSON-RPC).
type RawCommand struct{ Data interface{} }

func (RawCommand) isCommand() {}

func (r RawCommand) MarshalJSON() ([]byte, error) { return json.Marshal(r.Data) }

// RawInput holds arbitrary input JSON for pass-through (e.g. from JSON-RPC).
type RawInput struct{ Data interface{} }

func (RawInput) isInput() {}

func (r RawInput) MarshalJSON() ([]byte, error) { return json.Marshal(r.Data) }

// Argument is a command argument; use type assertion to distinguish.
// Implementations: *GasCoinArg, *InputArg, *ResultArg
type Argument interface {
	isArgument()
}

// GasCoinArg references the gas coin.
type GasCoinArg struct{}

func (GasCoinArg) isArgument() {}

// InputArg references an input index.
type InputArg struct {
	Index int `json:"index"`
}

func (InputArg) isArgument() {}

// ResultArg references a previous command result.
type ResultArg struct {
	Result    int  `json:"result"`
	Subresult *int `json:"subresult,omitempty"`
}

func (ResultArg) isArgument() {}

// Input is a programmable transaction input; use type assertion to distinguish.
// Implementations: various *Pure*Input, *SharedObjectInput, *ImmutableOrOwnedInput, *ReceivingInput
type Input interface {
	isInput()
}

// Pure input types (corresponding to Move primitive types).
// Supported: u8, u16, u32, u64, u128, u256, bool, address, string, object_id,
// vector<T>, std::option::Option<T> (T is a valid pure type, recursive).
// See https://docs.sui.io/concepts/transactions/inputs-and-results

// PureBCSInput holds raw BCS bytes (base64); used when the concrete type cannot be inferred (e.g. from gRPC pure field).
type PureBCSInput struct {
	Bytes string `json:"bytes"`
}

func (*PureBCSInput) isInput() {}

// PureU8Input is u8 type.
type PureU8Input struct {
	Value uint8 `json:"value"`
}

func (*PureU8Input) isInput() {}

// PureU16Input is u16 type.
type PureU16Input struct {
	Value uint16 `json:"value"`
}

func (*PureU16Input) isInput() {}

// PureU32Input u32 type
type PureU32Input struct {
	Value uint32 `json:"value"`
}

func (*PureU32Input) isInput() {}

// PureU64Input is u64 type.
type PureU64Input struct {
	Value uint64 `json:"value"`
}

func (*PureU64Input) isInput() {}

// PureU128Input is u128 type, stored as string to avoid overflow.
type PureU128Input struct {
	Value string `json:"value"`
}

func (*PureU128Input) isInput() {}

// PureU256Input is u256 type, stored as string.
type PureU256Input struct {
	Value string `json:"value"`
}

func (*PureU256Input) isInput() {}

// PureBoolInput is bool type.
type PureBoolInput struct {
	Value bool `json:"value"`
}

func (*PureBoolInput) isInput() {}

// PureAddressInput is address type.
type PureAddressInput struct {
	Value string `json:"value"`
}

func (*PureAddressInput) isInput() {}

// PureStringInput is string type (ASCII or UTF8).
type PureStringInput struct {
	Value string `json:"value"`
}

func (*PureStringInput) isInput() {}

// PureObjectIdInput is sui::object::ID type.
type PureObjectIdInput struct {
	Value string `json:"value"`
}

func (*PureObjectIdInput) isInput() {}

// PureVecInput is vector<T> type; T is a valid pure type (recursive).
type PureVecInput struct {
	Elements []Input `json:"elements"`
}

func (*PureVecInput) isInput() {}

type PureOptionInput struct {
	Value Input `json:"value,omitempty"`
}

func (*PureOptionInput) isInput() {}

// PureLiteralInput is a generic literal (from proto literal field); used when it cannot be mapped to a concrete type.
type PureLiteralInput struct {
	Value interface{} `json:"value"`
}

func (*PureLiteralInput) isInput() {}

// SharedObjectInput is a shared object input.
type SharedObjectInput struct {
	ObjectId             string `json:"objectId"`
	InitialSharedVersion string `json:"initialSharedVersion"`
	Mutable              bool   `json:"mutable"`
}

func (*SharedObjectInput) isInput() {}

// ImmutableOrOwnedInput is an immutable or owned object input.
type ImmutableOrOwnedInput struct {
	ObjectId string `json:"objectId"`
	Version  string `json:"version"`
	Digest   string `json:"digest"`
}

func (*ImmutableOrOwnedInput) isInput() {}

// ReceivingInput is a receiving object input.
type ReceivingInput struct {
	ObjectId string `json:"objectId"`
	Version  string `json:"version"`
	Digest   string `json:"digest"`
}

func (*ReceivingInput) isInput() {}

// Transaction is an executed (or simulated) Sui transaction with optional
// fields controlled by TransactionInclude.
type Transaction struct {
	Digest     string          `json:"digest"`
	Signatures []string        `json:"signatures"`
	Status     ExecutionStatus `json:"status"`
	Epoch      uint64          `json:"epoch"`
	Timestamp  int64           `json:"timestamp"`
	Checkpoint uint64          `json:"checkpoint"`

	// Populated when TransactionInclude.BalanceChanges == true.
	BalanceChanges []BalanceChange `json:"balanceChanges,omitempty"`
	// Populated when TransactionInclude.Effects == true.
	Effects *TransactionEffects `json:"effects,omitempty"`
	// Populated when TransactionInclude.Events == true.
	Events []Event `json:"events,omitempty"`
	// Populated when TransactionInclude.ObjectTypes == true.
	// Maps objectId → Move type string for all changed objects.
	ObjectTypes map[string]string `json:"objectTypes,omitempty"`
	// Populated when TransactionInclude.Transaction == true.
	// Parsed transaction data (sender, gas config, PTB inputs/commands).
	TransactionData *TransactionData `json:"transaction,omitempty"`
	// Populated when TransactionInclude.Bcs == true.
	// Raw BCS-encoded transaction bytes.
	BCS []byte `json:"bcs,omitempty"`
}

// TransactionResult is the discriminated union returned by GetTransaction and
// ExecuteTransaction.
type TransactionResult struct {
	Transaction *Transaction `json:"transaction,omitempty"`
}

// SimulateTransactionResult extends TransactionResult with command-level results
// (only available from SimulateTransaction).
type SimulateTransactionResult struct {
	Transaction *Transaction `json:"transaction,omitempty"`
	// Populated when SimulateTransactionInclude.CommandResults == true.
	CommandResults []CommandResult `json:"commandResults,omitempty"`
}

// GetTransactionOptions is the request for GetTransaction.
type GetTransactionOptions struct {
	Digest  string             `json:"digest"`
	Include TransactionInclude `json:"include,omitempty"`
}

// ExecuteTransactionOptions is the request for ExecuteTransaction.
type ExecuteTransactionOptions struct {
	// BCS-encoded transaction bytes.
	Transaction []byte             `json:"transaction"`
	Signatures  []string           `json:"signatures"`
	Include     TransactionInclude `json:"include,omitempty"`
}

// SimulateTransactionOptions is the request for SimulateTransaction.
type SimulateTransactionOptions struct {
	// BCS-encoded transaction bytes.
	Transaction []byte                     `json:"transaction"`
	Include     SimulateTransactionInclude `json:"include,omitempty"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// System state
// ═══════════════════════════════════════════════════════════════════════════════

// SystemParameters holds epoch-level protocol parameters.
type SystemParameters struct {
	EpochDurationMs              string `json:"epochDurationMs"`
	StakeSubsidyStartEpoch       string `json:"stakeSubsidyStartEpoch"`
	MaxValidatorCount            string `json:"maxValidatorCount"`
	MinValidatorJoiningStake     string `json:"minValidatorJoiningStake"`
	ValidatorLowStakeThreshold   string `json:"validatorLowStakeThreshold"`
	ValidatorLowStakeGracePeriod string `json:"validatorLowStakeGracePeriod"`
}

// StorageFund describes the on-chain storage fund balance.
type StorageFund struct {
	TotalObjectStorageRebates string `json:"totalObjectStorageRebates"`
	NonRefundableBalance      string `json:"nonRefundableBalance"`
}

// StakeSubsidy describes the subsidy schedule configuration.
type StakeSubsidy struct {
	Balance                   string `json:"balance"`
	DistributionCounter       string `json:"distributionCounter"`
	CurrentDistributionAmount string `json:"currentDistributionAmount"`
	StakeSubsidyPeriodLength  string `json:"stakeSubsidyPeriodLength"`
	StakeSubsidyDecreaseRate  int    `json:"stakeSubsidyDecreaseRate"`
}

// SystemStateInfo holds the core fields of the current SuiSystemState object.
type SystemStateInfo struct {
	SystemStateVersion              string           `json:"systemStateVersion"`
	Epoch                           string           `json:"epoch"`
	ProtocolVersion                 string           `json:"protocolVersion"`
	ReferenceGasPrice               string           `json:"referenceGasPrice"`
	EpochStartTimestampMs           string           `json:"epochStartTimestampMs"`
	SafeMode                        bool             `json:"safeMode"`
	SafeModeStorageRewards          string           `json:"safeModeStorageRewards"`
	SafeModeComputationRewards      string           `json:"safeModeComputationRewards"`
	SafeModeStorageRebates          string           `json:"safeModeStorageRebates"`
	SafeModeNonRefundableStorageFee string           `json:"safeModeNonRefundableStorageFee"`
	Parameters                      SystemParameters `json:"parameters"`
	StorageFund                     StorageFund      `json:"storageFund"`
	StakeSubsidy                    StakeSubsidy     `json:"stakeSubsidy"`
}

// GetCurrentSystemStateResponse is the response for GetCurrentSystemState.
type GetCurrentSystemStateResponse struct {
	SystemState SystemStateInfo `json:"systemState"`
}

// GetReferenceGasPriceResponse is the response for GetReferenceGasPrice.
type GetReferenceGasPriceResponse struct {
	ReferenceGasPrice string `json:"referenceGasPrice"`
}

// GetChainIdentifierResponse is the response for GetChainIdentifier.
// ChainIdentifier is the base58-encoded genesis checkpoint digest.
type GetChainIdentifierResponse struct {
	ChainIdentifier string `json:"chainIdentifier"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// ZkLogin
// ═══════════════════════════════════════════════════════════════════════════════

// ZkLoginIntentScope specifies the intent of the message being verified.
type ZkLoginIntentScope string

const (
	ZkLoginIntentScopeTransactionData ZkLoginIntentScope = "TransactionData"
	ZkLoginIntentScopePersonalMessage ZkLoginIntentScope = "PersonalMessage"
)

// VerifyZkLoginSignatureOptions is the request for VerifyZkLoginSignature.
type VerifyZkLoginSignatureOptions struct {
	// Base64-encoded message bytes.
	Bytes string `json:"bytes"`
	// Base64-encoded signature bytes.
	Signature   string             `json:"signature"`
	IntentScope ZkLoginIntentScope `json:"intentScope"`
	Address     string             `json:"address"`
}

// ZkLoginVerifyResponse is the response for VerifyZkLoginSignature.
type ZkLoginVerifyResponse struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// Name service
// ═══════════════════════════════════════════════════════════════════════════════

// DefaultNameServiceNameOptions is the request for DefaultNameServiceName.
type DefaultNameServiceNameOptions struct {
	Address string `json:"address"`
}

// DefaultNameServiceNameData holds the primary .sui name for an address.
type DefaultNameServiceNameData struct {
	// nil when the address has no registered .sui name.
	Name                *string `json:"name"`
	RegistrationNFTId   string  `json:"registrationNFTId"`
	ExpirationTimestamp string  `json:"expirationTimestamp"`
	TargetAddress       string  `json:"targetAddress"`
}

// DefaultNameServiceNameResponse is the response for DefaultNameServiceName.
type DefaultNameServiceNameResponse struct {
	Data DefaultNameServiceNameData `json:"data"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// Move package
// ═══════════════════════════════════════════════════════════════════════════════

// Visibility enumerates function access levels.
type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilityFriend  Visibility = "friend"
	VisibilityPrivate Visibility = "private"
	VisibilityUnknown Visibility = "unknown"
)

// Ability enumerates Move type abilities.
type Ability string

const (
	AbilityCopy    Ability = "copy"
	AbilityDrop    Ability = "drop"
	AbilityStore   Ability = "store"
	AbilityKey     Ability = "key"
	AbilityUnknown Ability = "unknown"
)

// TypeParameter represents a single Move type parameter with its constraints.
type TypeParameter struct {
	Constraints []Ability `json:"constraints"`
	IsPhantom   bool      `json:"isPhantom"`
}

// ReferenceType describes whether an OpenSignature argument is a reference.
type ReferenceType string

const (
	ReferenceTypeMutable   ReferenceType = "mutable"
	ReferenceTypeImmutable ReferenceType = "immutable"
	ReferenceTypeUnknown   ReferenceType = "unknown"
)

// OpenSignatureBodyKind enumerates the variants of an OpenSignature body.
type OpenSignatureBodyKind string

const (
	OpenSignatureBodyKindU8            OpenSignatureBodyKind = "u8"
	OpenSignatureBodyKindU16           OpenSignatureBodyKind = "u16"
	OpenSignatureBodyKindU32           OpenSignatureBodyKind = "u32"
	OpenSignatureBodyKindU64           OpenSignatureBodyKind = "u64"
	OpenSignatureBodyKindU128          OpenSignatureBodyKind = "u128"
	OpenSignatureBodyKindU256          OpenSignatureBodyKind = "u256"
	OpenSignatureBodyKindBool          OpenSignatureBodyKind = "bool"
	OpenSignatureBodyKindAddress       OpenSignatureBodyKind = "address"
	OpenSignatureBodyKindVector        OpenSignatureBodyKind = "vector"
	OpenSignatureBodyKindDatatype      OpenSignatureBodyKind = "datatype"
	OpenSignatureBodyKindTypeParameter OpenSignatureBodyKind = "typeParameter"
	OpenSignatureBodyKindUnknown       OpenSignatureBodyKind = "unknown"
)

// OpenSignatureBody is a recursive type that describes a Move type signature.
type OpenSignatureBody struct {
	Kind OpenSignatureBodyKind `json:"$kind"`
	// Populated for Kind == "vector": the element type.
	Vector *OpenSignatureBody `json:"vector,omitempty"`
	// Populated for Kind == "datatype".
	Datatype *DatatypeSignature `json:"datatype,omitempty"`
	// Populated for Kind == "typeParameter": the type parameter index.
	TypeParameterIndex int `json:"index"`
}

// DatatypeSignature holds the type name and type arguments for a datatype reference.
type DatatypeSignature struct {
	TypeName       string              `json:"typeName"`
	TypeParameters []OpenSignatureBody `json:"typeParameters"`
}

// OpenSignature describes one parameter or return value of a Move function,
// including its optional reference modifier.
type OpenSignature struct {
	// nil when the parameter is passed by value.
	Reference *ReferenceType    `json:"reference"`
	Body      OpenSignatureBody `json:"body"`
}

// FieldDescriptor describes one field in a Move struct or enum variant.
type FieldDescriptor struct {
	Name     string            `json:"name"`
	Position int               `json:"position"`
	Type     OpenSignatureBody `json:"type"`
}

// VariantDescriptor describes one variant of a Move enum.
type VariantDescriptor struct {
	Name     string            `json:"name"`
	Position int               `json:"position"`
	Fields   []FieldDescriptor `json:"fields"`
}

// DatatypeKind distinguishes structs from enums.
type DatatypeKind string

const (
	DatatypeKindStruct  DatatypeKind = "struct"
	DatatypeKindEnum    DatatypeKind = "enum"
	DatatypeKindUnknown DatatypeKind = "unknown"
)

// DatatypeResponse is a Move struct or enum datatype descriptor.
// Inspect Kind to determine which of Struct or Enum data is populated.
type DatatypeResponse struct {
	TypeName       string          `json:"typeName"`
	DefiningId     string          `json:"definingId"`
	ModuleName     string          `json:"moduleName"`
	Name           string          `json:"name"`
	Abilities      []Ability       `json:"abilities"`
	TypeParameters []TypeParameter `json:"typeParameters"`
	// Populated for DatatypeKindStruct.
	Fields []FieldDescriptor `json:"fields,omitempty"`
	// Populated for DatatypeKindEnum.
	Variants []VariantDescriptor `json:"variants,omitempty"`
}

// FunctionResponse is the descriptor for a Move function.
type FunctionResponse struct {
	PackageId      string          `json:"packageId"`
	ModuleName     string          `json:"moduleName"`
	Name           string          `json:"name"`
	Visibility     Visibility      `json:"visibility"`
	IsEntry        bool            `json:"isEntry"`
	TypeParameters []TypeParameter `json:"typeParameters"`
	Parameters     []OpenSignature `json:"parameters"`
	Returns        []OpenSignature `json:"returns"`
}

// GetMoveFunctionOptions is the request for GetMoveFunction.
type GetMoveFunctionOptions struct {
	PackageId  string `json:"packageId"`
	ModuleName string `json:"moduleName"`
	Name       string `json:"name"`
}

// GetMoveFunctionResponse is the response for GetMoveFunction.
type GetMoveFunctionResponse struct {
	Function FunctionResponse `json:"function"`
}
