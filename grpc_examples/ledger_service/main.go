package main

import (
	"context"
	"fmt"
	"log"

	"github.com/block-vision/sui-go-sdk/grpc_examples/utils"
	v2 "github.com/block-vision/sui-go-sdk/pb/sui/rpc/v2"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func main() {
	fmt.Println("=== Sui gRPC Ledger Service Examples ===")

	// Create authenticated gRPC client using common utility
	client := utils.CreateGrpcClientWithDefaults()
	defer client.Close()

	ctx := context.Background()

	// Get ledger service
	ledgerService, err := client.LedgerService(ctx)
	if err != nil {
		log.Fatalf("Failed to get ledger service: %v", err)
	}

	// Run all examples
	fmt.Println("\n1. Getting service info...")
	exampleGetServiceInfo(ctx, ledgerService)

	fmt.Println("\n2. Getting object...")
	exampleGetObject(ctx, ledgerService)

	fmt.Println("\n3. Batch getting objects...")
	exampleBatchGetObjects(ctx, ledgerService)

	fmt.Println("\n4. Getting transaction...")
	exampleGetTransaction(ctx, ledgerService)

	fmt.Println("\n5. Batch getting transactions...")
	exampleBatchGetTransactions(ctx, ledgerService)

	fmt.Println("\n6. Getting checkpoint...")
	exampleGetCheckpoint(ctx, ledgerService)

	fmt.Println("\n7. Getting epoch...")
	exampleGetEpoch(ctx, ledgerService)
}

// GetServiceInfo - Query the service for general information about its current state
func exampleGetServiceInfo(ctx context.Context, service v2.LedgerServiceClient) {
	req := &v2.GetServiceInfoRequest{}

	resp, err := service.GetServiceInfo(ctx, req)
	if err != nil {
		fmt.Printf("❌ GetServiceInfo failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Service Info: %v\n", resp)
}

// GetObject - Get a specific object by its ID
func exampleGetObject(ctx context.Context, service v2.LedgerServiceClient) {
	objectId := "0x88683c72e030b07af3881a005f376c2af1c30f7eeb99719c29b9ba5f151d8255"
	req := &v2.GetObjectRequest{
		ObjectId: &objectId,
		ReadMask: &fieldmaskpb.FieldMask{
			Paths: []string{"*"}, // Get all fields
		},
	}

	resp, err := service.GetObject(ctx, req)
	if err != nil {
		fmt.Printf("❌ GetObject failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Object: %v\n", resp.Object)
}

// BatchGetObjects - Get multiple objects by their IDs
func exampleBatchGetObjects(ctx context.Context, service v2.LedgerServiceClient) {
	objectId1 := "0x88683c72e030b07af3881a005f376c2af1c30f7eeb99719c29b9ba5f151d8255"
	objectId2 := "0x99694c72e030b07af3881a005f376c2af1c30f7eeb99719c29b9ba5f151d8256"

	req := &v2.BatchGetObjectsRequest{
		Requests: []*v2.GetObjectRequest{
			{ObjectId: &objectId1},
			{ObjectId: &objectId2},
		},
		ReadMask: &fieldmaskpb.FieldMask{
			Paths: []string{"object_id", "version", "digest"},
		},
	}

	resp, err := service.BatchGetObjects(ctx, req)
	if err != nil {
		fmt.Printf("❌ BatchGetObjects failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Batch Objects: Found %d objects\n", len(resp.Objects))
}

// GetTransaction - Get a specific transaction by its digest
func exampleGetTransaction(ctx context.Context, service v2.LedgerServiceClient) {
	txDigest := "3CK7Fv9CUp3QetDhszToqwnDHgzpYPauP1H2N5iyuijh"
	req := &v2.GetTransactionRequest{
		Digest: &txDigest,
		ReadMask: &fieldmaskpb.FieldMask{
			Paths: []string{"*"},
		},
	}

	resp, err := service.GetTransaction(ctx, req)
	if err != nil {
		fmt.Printf("❌ GetTransaction failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Transaction: %v\n", resp.Transaction)
}

// BatchGetTransactions - Get multiple transactions by their digests
func exampleBatchGetTransactions(ctx context.Context, service v2.LedgerServiceClient) {
	digest1 := "3CK7Fv9CUp3QetDhszToqwnDHgzpYPauP1H2N5iyuijh"
	digest2 := "29PhqkV8MkkDVDk2HLqsRZEUwTg7NgCgHvW7nVMUx2Xq"

	req := &v2.BatchGetTransactionsRequest{
		Digests: []string{digest1, digest2},
		ReadMask: &fieldmaskpb.FieldMask{
			Paths: []string{"digest", "transaction"},
		},
	}

	resp, err := service.BatchGetTransactions(ctx, req)
	if err != nil {
		fmt.Printf("❌ BatchGetTransactions failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Batch Transactions: Found %d transactions\n", len(resp.Transactions))
}

// GetCheckpoint - Get a specific checkpoint by sequence number
func exampleGetCheckpoint(ctx context.Context, service v2.LedgerServiceClient) {
	checkpointSeq := uint64(179720286)
	req := &v2.GetCheckpointRequest{
		CheckpointId: &v2.GetCheckpointRequest_SequenceNumber{
			SequenceNumber: checkpointSeq,
		},
		ReadMask: &fieldmaskpb.FieldMask{
			Paths: []string{"*"},
		},
	}

	resp, err := service.GetCheckpoint(ctx, req)
	if err != nil {
		fmt.Printf("❌ GetCheckpoint failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Checkpoint: %v\n", resp.Checkpoint)
}

// GetEpoch - Get epoch information
func exampleGetEpoch(ctx context.Context, service v2.LedgerServiceClient) {
	epochNumber := uint64(820)
	req := &v2.GetEpochRequest{
		Epoch: &epochNumber,
		ReadMask: &fieldmaskpb.FieldMask{
			Paths: []string{"*"},
		},
	}

	resp, err := service.GetEpoch(ctx, req)
	if err != nil {
		fmt.Printf("❌ GetEpoch failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Epoch: %v\n", resp.Epoch)
}
