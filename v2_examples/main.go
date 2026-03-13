// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// v2_example demonstrates how to use the sui/v2 high-level client built on gRPC.
package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/block-vision/sui-go-sdk/common/grpcconn"
	"github.com/block-vision/sui-go-sdk/common/httpconn"
	"github.com/block-vision/sui-go-sdk/constant"
	suiv2 "github.com/block-vision/sui-go-sdk/sui/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Supported method names
const (
	GetChainIdentifier     = "GetChainIdentifier"
	GetReferenceGasPrice   = "GetReferenceGasPrice"
	GetCurrentSystemState  = "GetCurrentSystemState"
	GetBalance             = "GetBalance"
	ListBalances           = "ListBalances"
	ListCoins              = "ListCoins"
	GetCoinMetadata        = "GetCoinMetadata"
	GetObjects             = "GetObjects"
	ListOwnedObjects       = "ListOwnedObjects"
	GetTransaction         = "GetTransaction"
	ListDynamicFields      = "ListDynamicFields"
	GetMoveFunction        = "GetMoveFunction"
	DefaultNameServiceName = "DefaultNameServiceName"
)

var supportedMethods = []string{
	GetChainIdentifier,
	GetReferenceGasPrice,
	GetCurrentSystemState,
	GetBalance,
	ListBalances,
	ListCoins,
	GetCoinMetadata,
	GetObjects,
	ListOwnedObjects,
	GetTransaction,
	ListDynamicFields,
	GetMoveFunction,
	DefaultNameServiceName,
}

const (
	testAddress  = "0xac5bceec1b789ff840d7d4e6ce4ce61c90d190a7f8c4f4ddf0bff6ee2413c33c"
	testObjectId = "0x88683c72e030b07af3881a005f376c2af1c30f7eeb99719c29b9ba5f151d8255"
	testTxDigest = "Bsq9m2aVYuv13BbSgGHqyjnobn87vhsv5LZQK2bkCzDo"
	testParentId = "0x266f5a401df5fa40fc5ab2a1a8e74ac41fe5fb241e106eb608bf37c732c17e0e"
)

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func main() {
	methodArg := ""
	if len(os.Args) >= 2 {
		methodArg = strings.TrimSpace(os.Args[1])
	}
	if methodArg == "-h" || methodArg == "--help" || methodArg == "help" || methodArg == "list" {
		fmt.Printf("Usage: %s [method]\n\n", os.Args[0])
		printSupportedMethods()
		fmt.Println("\nExamples:")
		fmt.Printf("  V2_TRANSPORT=grpc BLOCKVISION_API_KEY=... %s\n", os.Args[0])
		fmt.Printf("  V2_TRANSPORT=json-rpc %s GetBalance\n", os.Args[0])
		fmt.Printf("  %s list\n", os.Args[0])
		return
	}

	client, cleanup, transport, err := createClientFromEnv()
	if err != nil {
		log.Fatalf("Failed to create v2 client: %v", err)
	}
	defer cleanup()

	ctx := context.Background()

	runMethod := func(name string, fn func()) {
		if methodArg == "" || methodArg == "all" || strings.EqualFold(methodArg, name) {
			fmt.Printf("\n[%s]\n", name)
			fn()
		}
	}

	if methodArg != "" && methodArg != "all" {
		if !isSupportedMethod(methodArg) {
			fmt.Fprintf(os.Stderr, "Unknown method: %q\n\n", methodArg)
			printSupportedMethods()
			os.Exit(1)
		}
	}

	fmt.Printf("=== Sui v2 Client Examples (%s) ===\n", transport)

	runMethod(GetChainIdentifier, func() { exampleGetChainIdentifier(ctx, client) })
	runMethod(GetReferenceGasPrice, func() { exampleGetReferenceGasPrice(ctx, client) })
	runMethod(GetCurrentSystemState, func() { exampleGetCurrentSystemState(ctx, client) })
	runMethod(GetBalance, func() { exampleGetBalance(ctx, client) })
	runMethod(ListBalances, func() { exampleListBalances(ctx, client) })
	runMethod(ListCoins, func() { exampleListCoins(ctx, client) })
	runMethod(GetCoinMetadata, func() { exampleGetCoinMetadata(ctx, client) })
	runMethod(GetObjects, func() { exampleGetObjects(ctx, client) })
	runMethod(ListOwnedObjects, func() { exampleListOwnedObjects(ctx, client) })
	runMethod(GetTransaction, func() { exampleGetTransaction(ctx, client) })
	runMethod(ListDynamicFields, func() { exampleListDynamicFields(ctx, client) })
	runMethod(GetMoveFunction, func() { exampleGetMoveFunction(ctx, client) })
	runMethod(DefaultNameServiceName, func() { exampleDefaultNameServiceName(ctx, client) })
}

func createClientFromEnv() (suiv2.Client, func(), string, error) {
	transport := strings.ToLower(env("V2_TRANSPORT", "grpc"))
	switch transport {
	case "grpc":
		apiKey := os.Getenv("BLOCKVISION_API_KEY")
		if apiKey == "" {
			return nil, func() {}, transport, fmt.Errorf("BLOCKVISION_API_KEY is required for grpc transport")
		}
		grpcEndpoint := env("V2_GRPC_ENDPOINT", constant.SuiMainnetGrpcEndpoint)
		grpcClient := grpcconn.NewSuiGrpcClientWithAuth(
			grpcEndpoint,
			apiKey,
			grpcconn.WithTimeout(30*time.Second),
			grpcconn.WithRetryCount(3),
			grpcconn.WithDialOptions(
				grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
					InsecureSkipVerify: true,
				})),
				grpc.WithMaxMsgSize(20*1024*1024),
			),
		)
		client, err := suiv2.NewClient(suiv2.ClientOptions{
			GrpcClient: grpcClient,
		})
		if err != nil {
			grpcClient.Close()
			return nil, func() {}, transport, err
		}
		return client, func() { grpcClient.Close() }, transport, nil
	case "json-rpc", "jsonrpc", "http":
		rpcURL := env("V2_RPC_URL", constant.SuiMainnetEndpoint)
		httpConn := httpconn.NewHttpConn(rpcURL, nil)
		client, err := suiv2.NewClient(suiv2.ClientOptions{
			HttpConn: httpConn,
		})
		return client, func() {}, "json-rpc", err
	default:
		return nil, func() {}, transport, fmt.Errorf("unsupported V2_TRANSPORT %q", transport)
	}
}

func printSupportedMethods() {
	fmt.Println("Supported methods:")
	for i, m := range supportedMethods {
		fmt.Printf("  %2d. %s\n", i+1, m)
	}
}

func isSupportedMethod(name string) bool {
	for _, m := range supportedMethods {
		if strings.EqualFold(m, name) {
			return true
		}
	}
	return false
}

// ── helpers ──────────────────────────────────────────────────────────────────

func prettyPrint(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", v)
	}
	return string(b)
}

func printOK(label string, v interface{}) {
	fmt.Printf("  OK  %s:\n%s\n", label, prettyPrint(v))
}

func printErr(label string, err error) {
	fmt.Printf("  ERR %s: %v\n", label, err)
	var sdkErr *suiv2.SDKError
	if errors.As(err, &sdkErr) {
		fmt.Printf("      code=%s transport=%s method=%s operation=%s\n",
			sdkErr.Code, sdkErr.Transport, sdkErr.Method, sdkErr.Operation)
	}
}

// ── System / Chain ────────────────────────────────────────────────────────────

func exampleGetChainIdentifier(ctx context.Context, c suiv2.Client) {
	resp, err := c.GetChainIdentifier(ctx)
	if err != nil {
		printErr("GetChainIdentifier", err)
		return
	}
	printOK("ChainIdentifier", resp.ChainIdentifier)
}

func exampleGetReferenceGasPrice(ctx context.Context, c suiv2.Client) {
	resp, err := c.GetReferenceGasPrice(ctx)
	if err != nil {
		printErr("GetReferenceGasPrice", err)
		return
	}
	printOK("ReferenceGasPrice", resp.ReferenceGasPrice)
}

func exampleGetCurrentSystemState(ctx context.Context, c suiv2.Client) {
	resp, err := c.GetCurrentSystemState(ctx)
	if err != nil {
		printErr("GetCurrentSystemState", err)
		return
	}
	printOK("SystemState", resp.SystemState)
}

// ── Coins ──────────────────────────────────────────────────────────────────────

func exampleGetBalance(ctx context.Context, c suiv2.Client) {
	resp, err := c.GetBalance(ctx, suiv2.GetBalanceOptions{
		Owner: testAddress,
	})
	if err != nil {
		printErr("GetBalance", err)
		return
	}
	printOK("Balance", resp.Balance)
}

func exampleListBalances(ctx context.Context, c suiv2.Client) {
	resp, err := c.ListBalances(ctx, suiv2.ListBalancesOptions{
		Owner: testAddress,
	})
	if err != nil {
		printErr("ListBalances", err)
		return
	}
	fmt.Printf("  OK  ListBalances: %d balances, hasNextPage=%v\n", len(resp.Balances), resp.HasNextPage)
	for _, b := range resp.Balances {
		fmt.Printf("      coinType=%s  total=%s\n", b.CoinType, b.Balance_)
	}
}

func exampleListCoins(ctx context.Context, c suiv2.Client) {
	resp, err := c.ListCoins(ctx, suiv2.ListCoinsOptions{
		Owner: testAddress,
	})
	if err != nil {
		printErr("ListCoins", err)
		return
	}
	fmt.Printf("  OK  ListCoins: %d coins, hasNextPage=%v\n", len(resp.Objects), resp.HasNextPage)
	for _, coin := range resp.Objects {
		fmt.Printf("      id=%s  balance=%s\n", coin.ObjectId, coin.Balance)
	}
}

func exampleGetCoinMetadata(ctx context.Context, c suiv2.Client) {
	resp, err := c.GetCoinMetadata(ctx, suiv2.GetCoinMetadataOptions{
		CoinType: suiv2.SUI_TYPE_ARG,
	})
	if err != nil {
		printErr("GetCoinMetadata", err)
		return
	}
	if resp.CoinMetadata == nil {
		fmt.Println("  OK  GetCoinMetadata: no metadata returned")
		return
	}
	printOK("CoinMetadata", resp.CoinMetadata)
}

// ── Objects ────────────────────────────────────────────────────────────────────

func exampleGetObjects(ctx context.Context, c suiv2.Client) {
	resp, err := c.GetObjects(ctx, suiv2.GetObjectsOptions{
		ObjectIds: []string{testObjectId},
		Include: suiv2.ObjectInclude{
			PreviousTransaction: true,
			JSON:                true,
			// Content:             true,
			// ObjectBcs:           true,
		},
	})
	if err != nil {
		printErr("GetObjects", err)
		return
	}
	fmt.Printf("  OK  GetObjects: %d results\n", len(resp.Objects))
	for _, obj := range resp.Objects {
		if obj.Error != nil {
			fmt.Printf("      error: %s\n", *obj.Error)
		} else if obj.Object != nil {
			fmt.Printf("      id=%s  type=%s  version=%s\n", obj.Object.ObjectId, obj.Object.Type, obj.Object.Version)
		}
	}
}

func exampleListOwnedObjects(ctx context.Context, c suiv2.Client) {
	resp, err := c.ListOwnedObjects(ctx, suiv2.ListOwnedObjectsOptions{
		Owner: testAddress,
	})
	if err != nil {
		printErr("ListOwnedObjects", err)
		return
	}
	fmt.Printf("  OK  ListOwnedObjects: %d objects, hasNextPage=%v\n", len(resp.Objects), resp.HasNextPage)
	for _, obj := range resp.Objects {
		fmt.Printf("      id=%s  type=%s\n", obj.ObjectId, obj.Type)
	}
}

// ── Transactions ───────────────────────────────────────────────────────────────

func exampleGetTransaction(ctx context.Context, c suiv2.Client) {
	resp, err := c.GetTransaction(ctx, suiv2.GetTransactionOptions{
		Digest: testTxDigest,
		Include: suiv2.TransactionInclude{
			Transaction:    true,
			Effects:        true,
			BalanceChanges: true,
			Events:         true,
		},
	})
	if err != nil {
		printErr("GetTransaction", err)
		return
	}

	tx := resp.Transaction
	if tx == nil {
		fmt.Println("  OK  GetTransaction: no transaction payload returned")
		return
	}

	statusStr := "success"
	if !tx.Status.Success {
		statusStr = "failure"
	}
	fmt.Printf("  OK  GetTransaction: digest=%s  status=%s  epoch=%d\n",
		tx.Digest, statusStr, tx.Epoch)

	if tx.Effects != nil {
		fmt.Printf("      gasUsed: computation=%s  storage=%s  rebate=%s\n",
			tx.Effects.GasUsed.ComputationCost,
			tx.Effects.GasUsed.StorageCost,
			tx.Effects.GasUsed.StorageRebate,
		)
	}

	if tx.TransactionData != nil {
		for _, cmd := range tx.TransactionData.Commands {
			for k, v := range cmd {
				fmt.Printf("      command: %s\n", k)
				switch cmdValue := v.(type) {
				case *suiv2.MoveCallCommand:
					fmt.Printf("        package: %s\n", cmdValue.Package)
					fmt.Printf("        module: %s\n", cmdValue.Module)
					fmt.Printf("        function: %s\n", cmdValue.Function)
					fmt.Printf("        typeArguments: %v\n", cmdValue.TypeArguments)
					fmt.Printf("        arguments: %v\n", cmdValue.Arguments)
				default:
					fmt.Printf("        data: %s\n", prettyPrint(v))
				}
			}
		}
	}

	fmt.Printf("      balanceChanges: %d\n", len(tx.BalanceChanges))
	fmt.Printf("      events: %d\n", len(tx.Events))
}

// ── Other ──────────────────────────────────────────────────────────────────────

func exampleListDynamicFields(ctx context.Context, c suiv2.Client) {
	resp, err := c.ListDynamicFields(ctx, suiv2.ListDynamicFieldsOptions{
		ParentId: testParentId,
	})
	if err != nil {
		printErr("ListDynamicFields", err)
		return
	}
	fmt.Printf("  OK  ListDynamicFields: %d fields, hasNextPage=%v\n", len(resp.DynamicFields), resp.HasNextPage)
	for _, f := range resp.DynamicFields {
		fmt.Printf("      fieldId=%s  valueType=%s\n", f.FieldId, f.ValueType)
	}
}

func exampleGetMoveFunction(ctx context.Context, c suiv2.Client) {
	resp, err := c.GetMoveFunction(ctx, suiv2.GetMoveFunctionOptions{
		PackageId:  "0x2",
		ModuleName: "coin",
		Name:       "transfer",
	})
	if err != nil {
		printErr("GetMoveFunction", err)
		return
	}
	printOK("MoveFunction", resp.Function)
}

func exampleDefaultNameServiceName(ctx context.Context, c suiv2.Client) {
	resp, err := c.DefaultNameServiceName(ctx, suiv2.DefaultNameServiceNameOptions{
		Address: testAddress,
	})
	if err != nil {
		printErr("DefaultNameServiceName", err)
		return
	}
	if resp.Data.Name == nil {
		fmt.Println("  OK  DefaultNameServiceName: no .sui name registered for this address")
		return
	}
	printOK("DefaultNameServiceName", *resp.Data.Name)
}
