package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/block-vision/sui-go-sdk/common/grpcconn"
	"github.com/block-vision/sui-go-sdk/constant"
	"github.com/block-vision/sui-go-sdk/grpc_examples/utils"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/mystenbcs"
	v2 "github.com/block-vision/sui-go-sdk/pb/sui/rpc/v2"
	suisigner "github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/transaction"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func main() {
	fmt.Println("=== Sui gRPC Transaction Execution Service Examples ===")

	client := utils.CreateGrpcClientWithDefaults()
	defer client.Close()

	ctx := context.Background()

	txService, err := client.TransactionExecutionService(ctx)
	if err != nil {
		log.Fatalf("Failed to get transaction execution service: %v", err)
	}

	fmt.Println("\n1. Simulating transaction...")
	exampleSimulateTransaction(ctx, client, txService)

	fmt.Println("\n2. Executing transaction...")
	exampleExecuteTransaction(ctx, client, txService)
}

// exampleSimulateTransaction dry-runs a PTB without submitting it on-chain.
//
// Set SIMULATE_PRIVATE_KEY to a suiprivkey1... bech32-encoded key.
// The account must hold some SUI (testnet/devnet) for gas selection to work.
func exampleSimulateTransaction(ctx context.Context, client *grpcconn.SuiGrpcClient, txService v2.TransactionExecutionServiceClient) {
	privKey := os.Getenv("SIMULATE_PRIVATE_KEY")
	if privKey == "" {
		fmt.Println("  SKIP: set SIMULATE_PRIVATE_KEY to run this example")
		return
	}

	s, err := suisigner.NewSignerWithSecretKey(privKey)
	if err != nil {
		fmt.Printf("❌ NewSignerWithSecretKey: %v\n", err)
		return
	}

	gasPrice, err := fetchGasPrice(ctx, client)
	if err != nil {
		fmt.Printf("❌ fetchGasPrice: %v\n", err)
		return
	}

	tx := transaction.NewTransaction()
	tx.SetSigner(s)
	tx.SetGasPrice(gasPrice)

	coin := tx.Gas()
	splitResult := tx.SplitCoins(coin, []transaction.Argument{tx.Pure(uint64(1_000_000))})
	tx.TransferObjects([]transaction.Argument{splitResult}, tx.Pure(s.Address))

	// DoGasSelection: true — node picks gas coin automatically; Payment can be empty.
	bcsBytes, err := tx.BuildBCSBytes(ctx)
	if err != nil {
		fmt.Printf("❌ BuildBCSBytes: %v\n", err)
		return
	}

	doGasSelection := true
	resp, err := txService.SimulateTransaction(ctx, &v2.SimulateTransactionRequest{
		Transaction:    &v2.Transaction{Bcs: &v2.Bcs{Value: bcsBytes}},
		DoGasSelection: &doGasSelection,
		ReadMask: &fieldmaskpb.FieldMask{
			Paths: []string{
				"transaction.effects.status",
				"transaction.effects.gas_used",
			},
		},
	})
	if err != nil {
		fmt.Printf("❌ SimulateTransaction: %v\n", err)
		return
	}

	if resp.Transaction == nil {
		fmt.Println("❌ SimulateTransaction: empty response")
		return
	}

	status := "success"
	if effects := resp.Transaction.GetEffects(); effects != nil {
		if st := effects.GetStatus(); st != nil && !st.GetSuccess() {
			status = fmt.Sprintf("failure: %s", st.GetError())
		}
		if gu := effects.GetGasUsed(); gu != nil {
			fmt.Printf("✅ SimulateTransaction: status=%s  gasUsed: computation=%d storage=%d rebate=%d\n",
				status,
				gu.GetComputationCost(),
				gu.GetStorageCost(),
				gu.GetStorageRebate(),
			)
			return
		}
	}
	fmt.Printf("✅ SimulateTransaction: status=%s\n", status)
}

// exampleExecuteTransaction builds, signs, and submits a PTB on-chain.
//
// Set EXECUTE_PRIVATE_KEY to a suiprivkey1... bech32-encoded key.
// The account must hold SUI on testnet/devnet. The example transfers 1 MIST
// back to the sender (self-transfer) to keep the effect minimal.
func exampleExecuteTransaction(ctx context.Context, client *grpcconn.SuiGrpcClient, txService v2.TransactionExecutionServiceClient) {
	privKey := os.Getenv("EXECUTE_PRIVATE_KEY")
	if privKey == "" {
		fmt.Println("  SKIP: set EXECUTE_PRIVATE_KEY to run this example")
		return
	}

	s, err := suisigner.NewSignerWithSecretKey(privKey)
	if err != nil {
		fmt.Printf("❌ NewSignerWithSecretKey: %v\n", err)
		return
	}

	// Fetch current reference gas price.
	gasPrice, err := fetchGasPrice(ctx, client)
	if err != nil {
		fmt.Printf("❌ fetchGasPrice: %v\n", err)
		return
	}

	// Find a SUI coin to use as gas payment.
	gasCoin, err := fetchFirstSuiCoin(ctx, client, s.Address)
	if err != nil {
		fmt.Printf("❌ fetchFirstSuiCoin: %v\n", err)
		return
	}

	// Build PTB: split 1 MIST from gas coin, then transfer it back to self.
	tx := transaction.NewTransaction()
	tx.SetSigner(s)
	tx.SetGasPrice(gasPrice)
	tx.SetGasBudget(5_000_000)
	tx.SetGasPayment([]transaction.SuiObjectRef{*gasCoin})

	coin := tx.Gas()
	splitResult := tx.SplitCoins(coin, []transaction.Argument{tx.Pure(uint64(1))})
	tx.TransferObjects([]transaction.Argument{splitResult}, tx.Pure(s.Address))

	// Serialize to BCS bytes.
	bcsBytes, err := tx.BuildBCSBytes(ctx)
	if err != nil {
		fmt.Printf("❌ BuildBCSBytes: %v\n", err)
		return
	}

	// Sign: SignMessage expects a base64-encoded input.
	b64TxBytes := mystenbcs.ToBase64(bcsBytes)
	signed, err := s.SignMessage(b64TxBytes, constant.TransactionDataIntentScope)
	if err != nil {
		fmt.Printf("❌ SignMessage: %v\n", err)
		return
	}
	sigBytes, err := base64.StdEncoding.DecodeString(signed.Signature)
	if err != nil {
		fmt.Printf("❌ decode signature: %v\n", err)
		return
	}

	resp, err := txService.ExecuteTransaction(ctx, &v2.ExecuteTransactionRequest{
		Transaction: &v2.Transaction{
			Bcs: &v2.Bcs{Value: bcsBytes},
		},
		Signatures: []*v2.UserSignature{
			{Bcs: &v2.Bcs{Value: sigBytes}},
		},
		ReadMask: &fieldmaskpb.FieldMask{
			Paths: []string{"digest", "effects.status", "effects.gas_used"},
		},
	})
	if err != nil {
		fmt.Printf("❌ ExecuteTransaction: %v\n", err)
		return
	}

	if resp.Transaction == nil {
		fmt.Println("❌ ExecuteTransaction: empty response")
		return
	}

	status := "success"
	if effects := resp.Transaction.GetEffects(); effects != nil {
		if st := effects.GetStatus(); st != nil && !st.GetSuccess() {
			status = fmt.Sprintf("failure: %s", st.GetError())
		}
	}
	fmt.Printf("✅ ExecuteTransaction: digest=%s  status=%s\n",
		resp.Transaction.GetDigest(), status)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// fetchGasPrice returns the current reference gas price via LedgerService.GetEpoch.
func fetchGasPrice(ctx context.Context, client *grpcconn.SuiGrpcClient) (uint64, error) {
	ledgerService, err := client.LedgerService(ctx)
	if err != nil {
		return 0, fmt.Errorf("LedgerService: %w", err)
	}
	resp, err := ledgerService.GetEpoch(ctx, &v2.GetEpochRequest{
		ReadMask: &fieldmaskpb.FieldMask{Paths: []string{"reference_gas_price"}},
	})
	if err != nil {
		return 0, fmt.Errorf("GetEpoch: %w", err)
	}
	return resp.GetEpoch().GetReferenceGasPrice(), nil
}

// fetchFirstSuiCoin returns the first SUI coin owned by the address, suitable
// as a gas payment object.
func fetchFirstSuiCoin(ctx context.Context, client *grpcconn.SuiGrpcClient, address string) (*transaction.SuiObjectRef, error) {
	stateService, err := client.StateService(ctx)
	if err != nil {
		return nil, fmt.Errorf("StateService: %w", err)
	}

	coinType := "0x2::coin::Coin<0x2::sui::SUI>"
	pageSize := uint32(1)
	resp, err := stateService.ListOwnedObjects(ctx, &v2.ListOwnedObjectsRequest{
		Owner:      &address,
		ObjectType: &coinType,
		PageSize:   &pageSize,
		ReadMask: &fieldmaskpb.FieldMask{
			Paths: []string{"object_id", "version", "digest"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ListOwnedObjects: %w", err)
	}
	if len(resp.Objects) == 0 {
		return nil, fmt.Errorf("no SUI coins found for address %s", address)
	}

	obj := resp.Objects[0]
	ref, err := transaction.NewSuiObjectRef(
		models.SuiAddress(obj.GetObjectId()),
		fmt.Sprintf("%d", obj.GetVersion()),
		models.ObjectDigest(obj.GetDigest()),
	)
	if err != nil {
		return nil, fmt.Errorf("NewSuiObjectRef: %w", err)
	}
	return ref, nil
}
