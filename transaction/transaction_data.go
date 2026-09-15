package transaction

import (
	"bytes"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/mystenbcs"
)

type TransactionData struct {
	V1 *TransactionDataV1
}

func (*TransactionData) IsBcsEnum() {}

func (td *TransactionData) Marshal() ([]byte, error) {
	bcsEncodedMsg := bytes.Buffer{}
	bcsEncoder := mystenbcs.NewEncoder(&bcsEncodedMsg)
	err := bcsEncoder.Encode(td)
	if err != nil {
		return nil, err
	}

	return bcsEncodedMsg.Bytes(), nil
}

// TransactionDataV1 https://github.com/MystenLabs/sui/blob/fb27c6c7166f5e4279d5fd1b2ebc5580ca0e81b2/crates/sui-types/src/transaction.rs#L1625
type TransactionDataV1 struct {
	Kind       *TransactionKind
	Sender     *models.SuiAddressBytes
	GasData    *GasData
	Expiration *TransactionExpiration `bcs:"optional"`
}

func (td *TransactionDataV1) AddCommand(command Command) (index uint16) {
	index = uint16(len(td.Kind.ProgrammableTransaction.Commands))
	td.Kind.ProgrammableTransaction.Commands = append(td.Kind.ProgrammableTransaction.Commands, &command)

	return index
}

func (td *TransactionDataV1) AddInput(input CallArg) Argument {
	index := uint16(len(td.Kind.ProgrammableTransaction.Inputs))
	td.Kind.ProgrammableTransaction.Inputs = append(td.Kind.ProgrammableTransaction.Inputs, &input)

	return Argument{
		Input: &index,
	}
}

func (td *TransactionDataV1) GetInputObjectIndex(address models.SuiAddress) *uint16 {
	addressBytes, err := ConvertSuiAddressStringToBytes(address)
	if err != nil {
		return nil
	}

	for i, input := range td.Kind.ProgrammableTransaction.Inputs {
		if input.Object == nil {
			continue
		}

		if input.Object.ImmOrOwnedObject != nil {
			objectId := input.Object.ImmOrOwnedObject.ObjectId
			if objectId.IsEqual(*addressBytes) {
				index := uint16(i)
				return &index
			}
		}
		if input.Object.SharedObject != nil {
			objectId := input.Object.SharedObject.ObjectId
			if objectId.IsEqual(*addressBytes) {
				index := uint16(i)
				return &index
			}
		}
		if input.Object.Receiving != nil {
			objectId := input.Object.Receiving.ObjectId
			if objectId.IsEqual(*addressBytes) {
				index := uint16(i)
				return &index
			}
		}
	}

	return nil
}

// GasData https://github.com/MystenLabs/sui/blob/fb27c6c7166f5e4279d5fd1b2ebc5580ca0e81b2/crates/sui-types/src/transaction.rs#L1600
type GasData struct {
	Payment *[]SuiObjectRef
	Owner   *models.SuiAddressBytes
	Price   *uint64
	Budget  *uint64
}

func (gd *GasData) IsAllSet() bool {
	if gd.Payment == nil || gd.Owner == nil || gd.Price == nil || gd.Budget == nil {
		return false
	}

	return true
}

// TransactionExpiration https://github.com/MystenLabs/sui/blob/fb27c6c7166f5e4279d5fd1b2ebc5580ca0e81b2/crates/sui-types/src/transaction.rs#L1608
// - None
// - Epoch
type TransactionExpiration struct {
	None  any
	Epoch *uint64
}

func (*TransactionExpiration) IsBcsEnum() {}

// ProgrammableTransaction https://github.com/MystenLabs/sui/blob/fb27c6c7166f5e4279d5fd1b2ebc5580ca0e81b2/crates/sui-types/src/transaction.rs#L702
type ProgrammableTransaction struct {
	Inputs   []*CallArg
	Commands []*Command
}

// TransactionKind https://github.com/MystenLabs/sui/blob/fb27c6c7166f5e4279d5fd1b2ebc5580ca0e81b2/crates/sui-types/src/transaction.rs#L303
// - ProgrammableTransaction
// - ChangeEpoch
// - Genesis
// - ConsensusCommitPrologue
type TransactionKind struct {
	ProgrammableTransaction *ProgrammableTransaction
	ChangeEpoch             any
	Genesis                 any
	ConsensusCommitPrologue any
}

func (*TransactionKind) IsBcsEnum() {}

func (tk *TransactionKind) Marshal() ([]byte, error) {
	bcsEncodedMsg := bytes.Buffer{}
	bcsEncoder := mystenbcs.NewEncoder(&bcsEncodedMsg)
	err := bcsEncoder.Encode(tk)
	if err != nil {
		return nil, err
	}

	return bcsEncodedMsg.Bytes(), nil
}

func (tk *TransactionKind) Build() (string, error) {
	bcsEncodedMsg, err := tk.Marshal()
	if err != nil {
		return "", err
	}
	bcsBase64 := mystenbcs.ToBase64(bcsEncodedMsg)
	return bcsBase64, nil
}

// CallArg https://github.com/MystenLabs/sui/blob/fb27c6c7166f5e4279d5fd1b2ebc5580ca0e81b2/crates/sui-types/src/transaction.rs#L80
// - Pure
// - Object
// - UnresolvedPure
// - UnresolvedObject
type CallArg struct {
	Pure             *Pure
	Object           *ObjectArg
	UnresolvedPure   *UnresolvedPure
	UnresolvedObject *UnresolvedObject
}

func (*CallArg) IsBcsEnum() {}

type Pure struct {
	Bytes []byte
}

type UnresolvedPure struct {
	Value any
}

type UnresolvedObject struct {
	ObjectId models.SuiAddressBytes
	// Version
	// Digest
	// InitialSharedVersion
}

// ObjectArg
// - ImmOrOwnedObject
// - SharedObject
// - Receiving
type ObjectArg struct {
	ImmOrOwnedObject *SuiObjectRef
	SharedObject     *SharedObjectRef
	Receiving        *SuiObjectRef
}

func (ObjectArg) IsBcsEnum() {}

// Command https://github.com/MystenLabs/sui/blob/fb27c6c7166f5e4279d5fd1b2ebc5580ca0e81b2/crates/sui-types/src/transaction.rs#L712
// - MoveCall
// - TransferObjects
// - SplitCoins
// - MergeCoins
// - Publish
// - MakeMoveVec
// - Upgrade
type Command struct {
	MoveCall        *ProgrammableMoveCall
	TransferObjects *TransferObjects
	SplitCoins      *SplitCoins
	MergeCoins      *MergeCoins
	Publish         *Publish
	MakeMoveVec     *MakeMoveVec
	Upgrade         *Upgrade
}

func (*Command) IsBcsEnum() {}

// ProgrammableMoveCall https://github.com/MystenLabs/sui/blob/fb27c6c7166f5e4279d5fd1b2ebc5580ca0e81b2/crates/sui-types/src/transaction.rs#L762
type ProgrammableMoveCall struct {
	Package       models.SuiAddressBytes
	Module        string
	Function      string
	TypeArguments []*TypeTag
	Arguments     []*Argument
}

type TransferObjects struct {
	Objects []*Argument
	Address *Argument
}

type SplitCoins struct {
	Coin   *Argument
	Amount []*Argument
}

type MergeCoins struct {
	Destination *Argument
	Sources     []*Argument
}

type Publish struct {
	Modules      [][]byte
	Dependencies []models.SuiAddressBytes
}

// MakeMoveVec https://github.com/MystenLabs/sui/blob/fb27c6c7166f5e4279d5fd1b2ebc5580ca0e81b2/crates/sui-types/src/transaction.rs#L731
// Type is an Option, it is nil when the vector holds objects.
type MakeMoveVec struct {
	Type     *string `bcs:"optional"`
	Elements []*Argument
}

type Upgrade struct {
	Modules      [][]byte
	Dependencies []models.SuiAddressBytes
	Package      models.SuiAddressBytes
	Ticket       *Argument
}

// Argument https://github.com/MystenLabs/sui/blob/fb27c6c7166f5e4279d5fd1b2ebc5580ca0e81b2/crates/sui-types/src/transaction.rs#L745
// - GasCoin
// - Input
// - Result
// - NestedResult
type Argument struct {
	GasCoin      any
	Input        *uint16
	Result       *uint16
	NestedResult *NestedResult
}

func (*Argument) IsBcsEnum() {}

type NestedResult struct {
	Index       uint16
	ResultIndex uint16
}

type SuiObjectRef struct {
	ObjectId models.SuiAddressBytes
	Version  uint64
	Digest   models.ObjectDigestBytes
}

type SharedObjectRef struct {
	ObjectId             models.SuiAddressBytes
	InitialSharedVersion uint64
	Mutable              bool
}

type StructTag struct {
	Address    models.SuiAddressBytes
	Module     string
	Name       string
	TypeParams []*TypeTag
}

// TypeTag https://github.com/MystenLabs/sui/blob/ece197ed5c414eb274f99afc52704664af8d0c38/external-crates/move/crates/move-core-types/src/language_storage.rs#L33
// Do not reorder the fields, it will break the bcs encoding
type TypeTag struct {
	Bool    *bool
	U8      *bool
	U128    *bool
	U256    *bool
	Address *bool
	Signer  *bool
	Vector  *TypeTag
	Struct  *StructTag
	U16     *bool
	U32     *bool
	U64     *bool
}

func (*TypeTag) IsBcsEnum() {}
