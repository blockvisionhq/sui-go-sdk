package sui_json_rpc_types

import "github.com/block-vision/sui-go-sdk/models/sui_types"

type SuiCertifiedTransaction struct {
	TransactionDigest string                  `json:"transactionDigest,omitempty"`
	Data              SuiTransactionData      `json:"data,omitempty"`
	TxSignature       string                  `json:"txSignature,omitempty"`
	AuthSignInfo      AuthorityQuorumSignInfo `json:"authSignInfo,omitempty"`
}

type SuiParsedTransactionResponse struct {
	Publish   SuiParsedPublishResponse   `json:"publish,omitempty"`
	MergeCoin SuiParsedMergeCoinResponse `json:"mergeCoin,omitempty"`
	SplitCoin SuiParsedSplitCoinResponse `json:"splitCoin,omitempty"`
}

type SuiParsedPublishResponse struct {
	Package sui_types.SuiObjectRef `json:"package"`
}

type SuiParsedMergeCoinResponse struct {
}

type SuiParsedSplitCoinResponse struct {
}

type SuiTransactionData struct {
	Transactions []SuiTransactionKind   `json:"transactions,omitempty"`
	Sender       string                 `json:"sender,omitempty"`
	GasPayment   sui_types.SuiObjectRef `json:"gasPayment,omitempty"`
	GasBudget    uint64                 `json:"gasBudget,omitempty"`
}

type AuthorityQuorumSignInfo struct {
	Epoch      uint64   `json:"epoch,omitempty"`
	Signature  []string `json:"signature,omitempty"`
	SignersMap []uint64 `json:"signers_map,omitempty"`
}

type SuiTransactionKind struct {
	TransferObject TransferObject `json:"transferObject,omitempty"`
	Publish        Publish        `json:"publish,omitempty"`
	Call           Call           `json:"call,omitempty"`
	TransferSui    TransferSui    `json:"transferSui,omitempty"`
	ChangeEpoch    ChangeEpoch    `json:"changeEpoch,omitempty"`
}

type TransferObject struct {
	Recipient string                 `json:"recipient,omitempty"`
	ObjectRef sui_types.SuiObjectRef `json:"objectRef,omitempty"`
}

type Publish struct {
	Modules [][]byte `json:"modules,omitempty"`
}

type Call struct {
	Package       sui_types.SuiObjectRef `json:"package"`
	Module        string                 `json:"module"`
	Function      string                 `json:"function"`
	TypeArguments []interface{}          `json:"typeArguments"`
	Arguments     []interface{}          `json:"arguments"`
}

type TransferSui struct {
	Recipient string `json:"recipient,omitempty"`
	Amount    uint64 `json:"amount,omitempty"`
}

type ChangeEpoch struct {
	Epoch             uint64 `json:"epoch"`
	StorageCharge     uint64 `json:"storageCharge"`
	ComputationCharge uint64 `json:"computationCharge"`
}

type SuiTransactionEffects struct {
	Status            SuiExecutionStatus       `json:"status"`
	GasUsed           SuiGasCostSummary        `json:"gasUsed"`
	ShareObjects      []sui_types.SuiObjectRef `json:"shareObjects,omitempty"`
	TransactionDigest string                   `json:"transactionDigest"`
	Created           []OwnedObjectRef         `json:"created,omitempty"`
	Mutated           []OwnedObjectRef         `json:"mutated,omitempty"`
	Unwrapped         []OwnedObjectRef         `json:"unwrapped,omitempty"`
	Deleted           []sui_types.SuiObjectRef `json:"deleted,omitempty"`
	Wrapped           []sui_types.SuiObjectRef `json:"wrapped,omitempty"`
	GasObject         OwnedObjectRef           `json:"gasObject,omitempty"`
	Events            []SuiEvent               `json:"events,omitempty"`
	Dependencies      []string                 `json:"dependencies,omitempty"`
}

type SuiGasCostSummary struct {
	ComputationCost uint64 `json:"computationCost"`
	StorageCost     uint64 `json:"storageCost"`
	StorageRebate   uint64 `json:"storageRebate"`
}

type SuiExecutionStatus struct {
	Status string `json:"status"`
}

type OwnedObjectRef struct {
	Owner     sui_types.Owner        `json:"owner,omitempty"`
	Reference sui_types.SuiObjectRef `json:"reference,omitempty"`
}

type SuiEvent struct {
	MoveEvent MoveEvent `json:"moveEvent,omitempty"`
	Publish   Publish   `json:"publish,omitempty"`
}

type MoveEvent struct {
	PackageID string `json:"packageID,omitempty"`
}

type SuiParsedMoveObject struct {
	DataType          string                 `json:"dataType"`
	Type              string                 `json:"type"`
	HasPublicTransfer bool                   `json:"has_public_transfer"`
	Fields            map[string]interface{} `json:"fields"`
	BcsBytes          []byte                 `json:"bcs_bytes,omitempty"`
}

type SuiObjectInfo struct {
	OwnedObjectRef `json:"owner"`
}

type SuiEventEnvelop struct {
	Timestamp uint64 `json:"timestamp"`
	TxDigest  string `json:"txDigest,omitempty"`
	SuiEvent
}

type SuiMoveModuleId struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

type SuiMoveNormalizedModule struct {
	FileFormatVersion uint32                 `json:"fileFormatVersion"`
	Address           string                 `json:"address"`
	Name              string                 `json:"name"`
	Friends           []SuiMoveModuleId      `json:"friends"`
	Structs           map[string]interface{} `json:"structs"`
	ExposedFunctions  map[string]interface{} `json:"exposedFunctions"`
}

type SuiMoveNormalizedStruct struct {
	Abilities      interface{}   `json:"abilities"`
	TypeParameters []interface{} `json:"typeParameters"`
	Fields         []interface{} `json:"fields"`
}

type SuiMoveNormalizedFunction struct {
	Visibility     interface{}   `json:"visibility"`
	IsEntry        bool          `json:"isEntry"`
	TypeParameters []interface{} `json:"typeParameters"`
	Parameters     []interface{} `json:"parameters"`
	Return_        []interface{} `json:"return"`
}
