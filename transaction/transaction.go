package transaction

import (
	"bytes"
	"context"
	"math"
	"strconv"

	"github.com/block-vision/sui-go-sdk/constant"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/mystenbcs"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/block-vision/sui-go-sdk/utils"
	"github.com/jinzhu/copier"
	"github.com/samber/lo"
)

type Transaction struct {
	Data            TransactionData
	Signer          *signer.Signer
	SponsoredSigner *signer.Signer
	SuiClient       sui.ISuiAPI
}

func NewTransaction() *Transaction {
	data := TransactionData{
		V1: &TransactionDataV1{},
	}
	data.V1.Kind = &TransactionKind{
		ProgrammableTransaction: &ProgrammableTransaction{},
	}
	data.V1.GasData = &GasData{}

	return &Transaction{
		Data: data,
	}
}

func (tx *Transaction) SetSigner(signer *signer.Signer) *Transaction {
	tx.Signer = signer

	return tx
}

func (tx *Transaction) SetSponsoredSigner(signer *signer.Signer) *Transaction {
	tx.SponsoredSigner = signer

	return tx
}

func (tx *Transaction) SetSuiClient(client sui.ISuiAPI) *Transaction {
	tx.SuiClient = client

	return tx
}

func (tx *Transaction) SetSender(sender models.SuiAddress) *Transaction {
	address := utils.NormalizeSuiAddress(string(sender))
	addressBytes, err := ConvertSuiAddressStringToBytes(address)
	if err != nil {
		panic(err)
	}
	tx.Data.V1.Sender = addressBytes

	return tx
}

func (tx *Transaction) SetSenderIfNotSet(sender models.SuiAddress) *Transaction {
	if tx.Data.V1.Sender == nil {
		tx.SetSender(sender)
	}

	return tx
}

func (tx *Transaction) SetExpiration(expiration TransactionExpiration) *Transaction {
	tx.Data.V1.Expiration = &expiration

	return tx
}

func (tx *Transaction) SetGasPayment(payment []SuiObjectRef) *Transaction {
	tx.Data.V1.GasData.Payment = &payment

	return tx
}

func (tx *Transaction) SetGasOwner(owner models.SuiAddress) *Transaction {
	addressBytes, err := ConvertSuiAddressStringToBytes(owner)
	if err != nil {
		panic(err)
	}
	tx.Data.V1.GasData.Owner = addressBytes

	return tx
}

func (tx *Transaction) SetGasPrice(price uint64) *Transaction {
	tx.Data.V1.GasData.Price = &price

	return tx
}

func (tx *Transaction) SetGasBudget(budget uint64) *Transaction {
	tx.Data.V1.GasData.Budget = &budget

	return tx
}

func (tx *Transaction) SetGasBudgetIfNotSet(budget uint64) *Transaction {
	if tx.Data.V1.GasData.Budget == nil {
		tx.Data.V1.GasData.Budget = &budget
	}

	return tx
}

func (tx *Transaction) Gas() Argument {
	return Argument{
		GasCoin: struct{}{},
	}
}

func (tx *Transaction) Add(command Command) Argument {
	index := tx.Data.V1.AddCommand(command)

	return createTransactionResult(index, nil)
}

func (tx *Transaction) SplitCoins(coin Argument, amount []Argument) Argument {
	return tx.Add(splitCoins(SplitCoins{
		Coin:   &coin,
		Amount: convertArgumentsToArgumentPtrs(amount),
	}))
}

func (tx *Transaction) MergeCoins(destination Argument, sources []Argument) Argument {
	return tx.Add(mergeCoins(MergeCoins{
		Destination: &destination,
		Sources:     convertArgumentsToArgumentPtrs(sources),
	}))
}

func (tx *Transaction) Publish(modules [][]byte, dependencies []models.SuiAddress) Argument {
	dependenciesAddress := make([]models.SuiAddressBytes, len(dependencies))
	for i, dependency := range dependencies {
		v, err := ConvertSuiAddressStringToBytes(dependency)
		if err != nil {
			panic(err)
		}
		dependenciesAddress[i] = *v
	}

	return tx.Add(publish(Publish{
		Modules:      modules,
		Dependencies: dependenciesAddress,
	}))
}

func (tx *Transaction) Upgrade(
	modules [][]byte,
	dependencies []models.SuiAddress,
	packageId models.SuiAddress,
	ticket Argument,
) Argument {
	dependenciesAddress := make([]models.SuiAddressBytes, len(dependencies))
	for i, dependency := range dependencies {
		v, err := ConvertSuiAddressStringToBytes(dependency)
		if err != nil {
			panic(err)
		}
		dependenciesAddress[i] = *v
	}

	packageIdBytes, err := ConvertSuiAddressStringToBytes(packageId)
	if err != nil {
		panic(err)
	}

	return tx.Add(upgrade(Upgrade{
		Modules:      modules,
		Dependencies: dependenciesAddress,
		Package:      *packageIdBytes,
		Ticket:       &ticket,
	}))
}

func (tx *Transaction) MoveCall(
	packageId models.SuiAddress,
	module string,
	function string,
	typeArguments []TypeTag,
	arguments []Argument,
) Argument {
	packageIdBytes, err := ConvertSuiAddressStringToBytes(packageId)
	if err != nil {
		panic(err)
	}

	return tx.Add(moveCall(ProgrammableMoveCall{
		Package:       *packageIdBytes,
		Module:        module,
		Function:      function,
		TypeArguments: convertTypeTagsToTypeTagPtrs(typeArguments),
		Arguments:     convertArgumentsToArgumentPtrs(arguments),
	}))
}

func (tx *Transaction) TransferObjects(objects []Argument, address Argument) Argument {
	return tx.Add(transferObjects(TransferObjects{
		Objects: convertArgumentsToArgumentPtrs(objects),
		Address: &address,
	}))
}

func (tx *Transaction) MakeMoveVec(typeValue *string, elements []Argument) Argument {
	return tx.Add(makeMoveVec(MakeMoveVec{
		Type:     typeValue,
		Elements: convertArgumentsToArgumentPtrs(elements),
	}))
}

// Object
// - input: string | CallArg | Argument
func (tx *Transaction) Object(input any) Argument {
	// string
	if s, ok := input.(string); ok {
		if utils.IsValidSuiAddress(models.SuiAddress(s)) {
			address := utils.NormalizeSuiAddress(s)
			addressBytes, err := ConvertSuiAddressStringToBytes(address)
			if err != nil {
				panic(err)
			}

			// TODO: Support more object types
			arg := tx.Data.V1.AddInput(CallArg{
				UnresolvedObject: &UnresolvedObject{
					ObjectId: *addressBytes,
				},
			})

			return arg
		} else {
			panic(ErrObjectNotSupportType)
		}
	}

	// Argument
	if arg, ok := input.(Argument); ok {
		return arg
	}

	// CallArg
	if v, ok := input.(CallArg); ok {
		isTypeSupported := false

		if v.Object.SharedObject != nil {
			// SharedObject
			address := ConvertSuiAddressBytesToString(v.Object.SharedObject.ObjectId)
			if index := tx.Data.V1.GetInputObjectIndex(address); index != nil {
				if v.Object.SharedObject.Mutable {
					newExistObject := tx.Data.V1.Kind.ProgrammableTransaction.Inputs[*index]
					if newExistObject.Object.SharedObject != nil {
						newExistObject.Object.SharedObject.Mutable = true
						tx.Data.V1.Kind.ProgrammableTransaction.Inputs[*index] = newExistObject
					}
				}

				return Argument{
					Input: index,
				}
			}

			isTypeSupported = true
		}

		if v.Object.ImmOrOwnedObject != nil {
			isTypeSupported = true
		}
		if v.Object.Receiving != nil {
			isTypeSupported = true
		}

		if isTypeSupported {
			arg := tx.Data.V1.AddInput(CallArg{
				Object: v.Object,
			})
			return arg
		}
	}

	panic(ErrObjectNotSupportType)
}

func (tx *Transaction) Pure(input any) Argument {
	var val []byte
	if s, ok := input.(string); ok && utils.IsValidSuiAddress(models.SuiAddress(s)) {
		fixedAddressBytes, err := ConvertSuiAddressStringToBytes(models.SuiAddress(s))
		if err != nil {
			panic(err)
		}
		addressBytes := fixedAddressBytes[:]
		val = addressBytes
	} else {
		bcsEncodedMsg := bytes.Buffer{}
		bcsEncoder := mystenbcs.NewEncoder(&bcsEncodedMsg)
		err := bcsEncoder.Encode(input)
		if err != nil {
			panic(err)
		}
		val = bcsEncodedMsg.Bytes()
	}

	arg := tx.Data.V1.AddInput(CallArg{Pure: &Pure{
		Bytes: val,
	}})

	return arg
}

// BuildBCSBytes resolves sender, gas price and gas budget (same as Execute),
// then returns the raw BCS-encoded TransactionData bytes suitable for
// SimulateTransactionOptions.Transaction or ExecuteTransactionOptions.Transaction.
//
// Unlike Execute, this method does NOT require gas payment to be set when
// DoGasSelection will be true on the gRPC SimulateTransaction call.
func (tx *Transaction) BuildBCSBytes(ctx context.Context) ([]byte, error) {
	if tx.Signer == nil {
		return nil, ErrSignerNotSet
	}

	if tx.Data.V1.GasData.Price == nil {
		if tx.SuiClient != nil {
			rsp, err := tx.SuiClient.SuiXGetReferenceGasPrice(ctx)
			if err != nil {
				return nil, err
			}
			tx.SetGasPrice(rsp)
		} else {
			return nil, ErrGasPriceNotSet
		}
	}
	tx.SetGasBudgetIfNotSet(defaultGasBudget)
	tx.SetSenderIfNotSet(models.SuiAddress(tx.Signer.Address))

	if tx.Data.V1.Sender == nil {
		return nil, ErrSenderNotSet
	}
	if tx.Data.V1.GasData.Owner == nil {
		tx.SetGasOwner(models.SuiAddress(tx.Signer.Address))
	}
	// DoGasSelection on the gRPC side replaces the payment at execution time,
	// but the node still needs to fully parse the BCS bytes first.
	// An empty slice encodes as a valid BCS vector (0x00); a nil pointer does not.
	if tx.Data.V1.GasData.Payment == nil {
		empty := []SuiObjectRef{}
		tx.Data.V1.GasData.Payment = &empty
	}

	return tx.Data.Marshal()
}

func (tx *Transaction) Execute(
	ctx context.Context,
	options models.SuiTransactionBlockOptions,
	requestType string,
) (*models.SuiTransactionBlockResponse, error) {
	if tx.SuiClient == nil {
		return nil, ErrSuiClientNotSet
	}
	req, err := tx.ToSuiExecuteTransactionBlockRequest(ctx, options, requestType)
	if err != nil {
		return nil, err
	}
	rsp, err := tx.SuiClient.SuiExecuteTransactionBlock(ctx, *req)
	if err != nil {
		return nil, err
	}

	return &rsp, nil
}

func (tx *Transaction) ToSuiExecuteTransactionBlockRequest(
	ctx context.Context,
	options models.SuiTransactionBlockOptions,
	requestType string,
) (*models.SuiExecuteTransactionBlockRequest, error) {
	if tx.Signer == nil {
		return nil, ErrSignerNotSet
	}

	b64TxBytes, err := tx.buildTransaction(ctx)
	if err != nil {
		return nil, err
	}
	var signatures []string
	if tx.SponsoredSigner != nil {
		sponsoredMessage, err := tx.SponsoredSigner.SignMessage(b64TxBytes, constant.TransactionDataIntentScope)
		if err != nil {
			return nil, err
		}
		signatures = append(signatures, sponsoredMessage.Signature)
	}
	message, err := tx.Signer.SignMessage(b64TxBytes, constant.TransactionDataIntentScope)
	if err != nil {
		return nil, err
	}
	signatures = append(signatures, message.Signature)

	return &models.SuiExecuteTransactionBlockRequest{
		TxBytes:     b64TxBytes,
		Signature:   signatures,
		Options:     options,
		RequestType: requestType,
	}, nil
}

func (tx *Transaction) buildTransaction(ctx context.Context) (string, error) {
	if tx.Signer == nil {
		return "", ErrSignerNotSet
	}

	if tx.Data.V1.GasData.Price == nil {
		if tx.SuiClient != nil {
			rsp, err := tx.SuiClient.SuiXGetReferenceGasPrice(ctx)
			if err != nil {
				return "", err
			}
			tx.SetGasPrice(rsp)
		}
	}
	tx.SetGasBudgetIfNotSet(defaultGasBudget)
	tx.SetSenderIfNotSet(models.SuiAddress(tx.Signer.Address))

	return tx.build(false)
}

func (tx *Transaction) build(onlyTransactionKind bool) (string, error) {
	if onlyTransactionKind {
		bcsEncodedMsg, err := tx.Data.V1.Kind.Marshal()
		if err != nil {
			return "", err
		}
		bcsBase64 := mystenbcs.ToBase64(bcsEncodedMsg)
		return bcsBase64, nil
	}

	if tx.Data.V1.Sender == nil {
		return "", ErrSenderNotSet
	}
	if tx.Data.V1.GasData.Owner == nil {
		tx.SetGasOwner(models.SuiAddress(tx.Signer.Address))
	}
	if !tx.Data.V1.GasData.IsAllSet() {
		return "", ErrGasDataNotAllSet
	}

	bcsEncodedMsg, err := tx.Data.Marshal()
	if err != nil {
		return "", err
	}
	bcsBase64 := mystenbcs.ToBase64(bcsEncodedMsg)

	return bcsBase64, nil
}

func (tx *Transaction) NewTransactionFromKind() (newTx *Transaction, err error) {
	newTx = NewTransaction()
	err = copier.CopyWithOption(&newTx.Data.V1.Kind, &tx.Data.V1.Kind, copier.Option{DeepCopy: true})
	if err != nil {
		return nil, err
	}
	return newTx, nil
}

func NewSuiObjectRef(objectId models.SuiAddress, version string, digest models.ObjectDigest) (*SuiObjectRef, error) {
	objectIdBytes, err := ConvertSuiAddressStringToBytes(objectId)
	if err != nil {
		return nil, err
	}
	digestBytes, err := ConvertObjectDigestStringToBytes(digest)
	if err != nil {
		return nil, err
	}
	versionUint64, err := strconv.ParseUint(version, 10, 64)
	if err != nil {
		return nil, err
	}

	return &SuiObjectRef{
		ObjectId: *objectIdBytes,
		Version:  versionUint64,
		Digest:   *digestBytes,
	}, nil
}

func createTransactionResult(index uint16, length *uint16) Argument {
	if length == nil {
		length = lo.ToPtr(uint16(math.MaxUint16))
	}

	// TODO: Support NestedResult
	return Argument{
		Result: lo.ToPtr(index),
	}
}

func convertArgumentsToArgumentPtrs(args []Argument) []*Argument {
	argPtrs := make([]*Argument, len(args))
	for i, arg := range args {
		v := arg
		argPtrs[i] = &v
	}

	return argPtrs
}

func convertTypeTagsToTypeTagPtrs(tags []TypeTag) []*TypeTag {
	tagPtrs := make([]*TypeTag, len(tags))
	for i, tag := range tags {
		v := tag
		tagPtrs[i] = &v
	}

	return tagPtrs
}
