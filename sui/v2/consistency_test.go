// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v2

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/block-vision/sui-go-sdk/common/grpcconn"
	"github.com/block-vision/sui-go-sdk/common/httpconn"
	"github.com/block-vision/sui-go-sdk/constant"
	"github.com/block-vision/sui-go-sdk/mystenbcs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

var (
	grpcEndpoint = env("GRPC_ENDPOINT", constant.SuiMainnetGrpcEndpoint)
	rpcURL       = env("RPC_URL", constant.SuiMainnetEndpoint)
	testAddress  = env("TEST_ADDRESS", "0x14f6dcefbbc67d4291c7cafe37e33f5391bf2f9c4519fe037bbb0a1e2dd5e2b6")
	testObjectId = env("TEST_OBJECT_ID", "0x88683c72e030b07af3881a005f376c2af1c30f7eeb99719c29b9ba5f151d8255")
	testTxDigest = env("TEST_TX_DIGEST", "GDaXJ2D69d763ZA6hN8wksosD3NGPShTvMWJGZYaDT1e")
	testParentId = env("TEST_PARENT_ID", "0x266f5a401df5fa40fc5ab2a1a8e74ac41fe5fb241e106eb608bf37c732c17e0e")
	testPkgId    = env("TEST_PACKAGE_ID", "0xeb9210e2980489154cc3c293432b9a1b1300edd0d580fe2269dd9cda34baee6d")
	testModule   = env("TEST_MODULE", "swap_router")
	testFunction = env("TEST_FUNCTION", "swap_a_b")
)

// setupClients creates gRPC and JSON RPC clients.
// If gRPC connection fails (e.g. no API key), grpcClient is nil and tests run JSON RPC only.
func setupClients(t *testing.T) (grpcClient Client, jsonClient Client) {
	t.Helper()

	// JSON RPC client - use a custom transport so tests are not blocked by local
	// trust-store issues when the upstream endpoint serves a non-standard chain.
	httpConn := httpconn.NewCustomHttpConn(rpcURL, &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	})
	jsonClient, err := NewClient(ClientOptions{HttpConn: httpConn})
	if err != nil {
		t.Fatalf("failed to create JSON RPC client: %v", err)
	}

	// gRPC client - optional; set BLOCKVISION_API_KEY or GRPC_TOKEN.
	apiKey := env("BLOCKVISION_API_KEY", env("GRPC_TOKEN", ""))
	if apiKey == "" {
		return nil, jsonClient
	}
	gc := grpcconn.NewSuiGrpcClientWithAuth(
		grpcEndpoint,
		apiKey,
		grpcconn.WithTimeout(30*time.Second),
		grpcconn.WithDialOptions(
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})),
			grpc.WithMaxMsgSize(20*1024*1024),
		),
	)
	grpcClient, err = NewClient(ClientOptions{GrpcClient: gc})
	if err != nil {
		t.Logf("gRPC client not available (skip consistency test): %v", err)
		gc.Close()
		return nil, jsonClient
	}
	t.Cleanup(func() { gc.Close() })
	return grpcClient, jsonClient
}

// assertResponseEqual compares JSON serialization of two responses.
// JSON comparison ignores pointer/nil/empty slice implementation differences.
func assertResponseEqual(t *testing.T, name string, a, b interface{}) bool {
	t.Helper()
	return assertResponseEqualEx(t, name, a, b, nil, nil)
}

func assertResponseEqualEx(t *testing.T, name string, a, b interface{}, ignorePaths []string, normalizer func(interface{}) interface{}) bool {
	t.Helper()
	ja, err := json.Marshal(a)
	if err != nil {
		t.Errorf("%s: marshal a failed: %v", name, err)
		return false
	}
	jb, err := json.Marshal(b)
	if err != nil {
		t.Errorf("%s: marshal b failed: %v", name, err)
		return false
	}
	var va, vb interface{}
	if err := json.Unmarshal(ja, &va); err != nil {
		t.Errorf("%s: unmarshal a failed: %v", name, err)
		return false
	}
	if err := json.Unmarshal(jb, &vb); err != nil {
		t.Errorf("%s: unmarshal b failed: %v", name, err)
		return false
	}
	if len(ignorePaths) > 0 {
		va = dropPaths(va, ignorePaths)
		vb = dropPaths(vb, ignorePaths)
	}
	if normalizer != nil {
		va = normalizer(va)
		vb = normalizer(vb)
	}
	ja2, _ := json.Marshal(va)
	jb2, _ := json.Marshal(vb)
	if string(ja2) != string(jb2) {
		t.Errorf("%s: response mismatch\n=== gRPC ===\n%s\n=== JSON RPC ===\n%s", name, string(ja2), string(jb2))
		return false
	}
	return true
}

func normalizeTxForCompare(obj interface{}) interface{} {
	m, ok := obj.(map[string]interface{})
	if !ok {
		return obj
	}
	tx, _ := m["transaction"].(map[string]interface{})
	if tx == nil {
		return obj
	}
	effects, _ := tx["effects"].(map[string]interface{})
	if effects != nil {
		for _, key := range []string{"changedObjects", "gasObject"} {
			switch v := effects[key].(type) {
			case []interface{}:
				for _, item := range v {
					if co, ok := item.(map[string]interface{}); ok {
						delete(co, "inputDigest")
					}
				}
				if key == "changedObjects" {
					sort.Slice(v, func(i, j int) bool {
						mi, _ := v[i].(map[string]interface{})
						mj, _ := v[j].(map[string]interface{})
						return stringifyForSort(mi["objectId"]) < stringifyForSort(mj["objectId"])
					})
				}
			case map[string]interface{}:
				delete(v, "inputDigest")
			}
		}
		delete(effects, "eventsDigest")
	}
	delete(tx, "events")
	normalizeTypeArgumentsInCommands(tx)
	normalizePureInputs(tx)
	return obj
}

var suiAddrShortRegex = regexp.MustCompile(`0x([0-9a-fA-F]+)(::|$)`)

func expandSuiAddressInType(s string) string {
	return suiAddrShortRegex.ReplaceAllStringFunc(s, func(m string) string {
		idx := strings.Index(m, "::")
		suffix := ""
		hexPart := m
		if idx >= 0 {
			suffix = m[idx:]
			hexPart = m[2:idx]
		} else {
			hexPart = m[2:]
		}
		if len(hexPart) >= 64 {
			return m
		}
		hexPart = strings.TrimLeft(hexPart, "0")
		if hexPart == "" {
			hexPart = "0"
		}
		if len(hexPart)%2 != 0 {
			hexPart = "0" + hexPart
		}
		for len(hexPart) < 64 {
			hexPart = "0" + hexPart
		}
		return "0x" + hexPart + suffix
	})
}

func normalizeTypeArgumentsInCommands(tx map[string]interface{}) {
	innerTx, _ := tx["transaction"].(map[string]interface{})
	if innerTx == nil {
		return
	}
	commands, _ := innerTx["commands"].([]interface{})
	for _, cmd := range commands {
		cmdMap, _ := cmd.(map[string]interface{})
		if cmdMap == nil {
			continue
		}
		for _, v := range cmdMap {
			if moveCall, ok := v.(map[string]interface{}); ok {
				if ta, ok := moveCall["typeArguments"].([]interface{}); ok {
					for i, item := range ta {
						if s, ok := item.(string); ok {
							ta[i] = expandSuiAddressInType(s)
						}
					}
				}
			}
		}
	}
}

func normalizePureInputs(tx map[string]interface{}) {
	innerTx, _ := tx["transaction"].(map[string]interface{})
	if innerTx == nil {
		return
	}
	inputs, _ := innerTx["inputs"].([]interface{})
	for i, input := range inputs {
		m, ok := input.(map[string]interface{})
		if !ok {
			continue
		}
		t, _ := m["type"].(string)
		if t != "pure" {
			continue
		}
		valueType, _ := m["valueType"].(string)
		b64 := pureValueToBytesNormalize(m["value"], valueType)
		if b64 != "" {
			inputs[i] = map[string]interface{}{"bytes": b64}
		}
	}
}

func normalizeMoveTypeString(s string) string {
	return strings.ReplaceAll(s, ", ", ",")
}

func normalizeMoveTypes(obj interface{}) interface{} {
	switch v := obj.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if key == "type" {
				if s, ok := value.(string); ok {
					v[key] = normalizeMoveTypeString(s)
					continue
				}
			}
			v[key] = normalizeMoveTypes(value)
		}
		return v
	case []interface{}:
		for i, item := range v {
			v[i] = normalizeMoveTypes(item)
		}
		return v
	default:
		return obj
	}
}

func pureValueToBytesNormalize(val interface{}, valueType string) string {
	if valueType == "" || !strings.HasPrefix(valueType, "vector<") {
		return ""
	}
	arr, ok := val.([]interface{})
	if !ok {
		return ""
	}
	innerType := strings.TrimSuffix(strings.TrimPrefix(valueType, "vector<"), ">")
	if innerType != "u64" {
		return ""
	}
	b := mystenbcs.ULEB128Encode(len(arr))
	for _, x := range arr {
		var u uint64
		switch v := x.(type) {
		case float64:
			u = uint64(v)
		case string:
			u, _ = strconv.ParseUint(v, 10, 64)
		default:
			return ""
		}
		eb := make([]byte, 8)
		binary.LittleEndian.PutUint64(eb, u)
		b = append(b, eb...)
	}
	return base64.StdEncoding.EncodeToString(b)
}

func dropPaths(obj interface{}, paths []string) interface{} {
	m, ok := obj.(map[string]interface{})
	if !ok {
		return obj
	}
	for _, path := range paths {
		dropPath(m, strings.Split(path, "."))
	}
	return m
}

func dropPath(m map[string]interface{}, parts []string) {
	if len(parts) == 0 {
		return
	}
	if len(parts) == 1 {
		delete(m, parts[0])
		return
	}
	child, ok := m[parts[0]]
	if !ok {
		return
	}
	switch c := child.(type) {
	case map[string]interface{}:
		dropPath(c, parts[1:])
	case []interface{}:
		for _, item := range c {
			if childMap, ok := item.(map[string]interface{}); ok {
				dropPath(childMap, parts[1:])
			}
		}
	}
}

func stringifyForSort(v interface{}) string {
	s, _ := v.(string)
	return s
}

func collectAllBalances(ctx context.Context, client Client, opts ListBalancesOptions) (*ListBalancesResponse, error) {
	all := &ListBalancesResponse{Balances: []Balance{}}
	for {
		resp, err := client.ListBalances(ctx, opts)
		if err != nil {
			return nil, err
		}
		all.Balances = append(all.Balances, resp.Balances...)
		if !resp.HasNextPage || resp.Cursor == nil {
			break
		}
		opts.Cursor = resp.Cursor
	}
	return all, nil
}

func collectAllOwnedObjects(ctx context.Context, client Client, opts ListOwnedObjectsOptions) (*ListOwnedObjectsResponse, error) {
	all := &ListOwnedObjectsResponse{Objects: []Object{}}
	for {
		resp, err := client.ListOwnedObjects(ctx, opts)
		if err != nil {
			return nil, err
		}
		all.Objects = append(all.Objects, resp.Objects...)
		if !resp.HasNextPage || resp.Cursor == nil {
			break
		}
		opts.Cursor = resp.Cursor
	}
	return all, nil
}

// TestResponseConsistency verifies JSON RPC and gRPC response consistency.
// Requires network, skip with -short: go test -short ./sui/v2/...
func TestResponseConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping consistency test in short mode")
	}
	grpcClient, jsonClient := setupClients(t)
	if grpcClient == nil {
		t.Skip("gRPC client not available, skip consistency test")
	}

	ctx := context.Background()

	t.Run("GetChainIdentifier", func(t *testing.T) {
		t.Skip("GetChainIdentifier: gRPC returns base58 genesis digest, JSON RPC returns hex chain id, semantics differ")
	})

	t.Run("GetReferenceGasPrice", func(t *testing.T) {
		gr, err1 := grpcClient.GetReferenceGasPrice(ctx)
		jr, err2 := jsonClient.GetReferenceGasPrice(ctx)
		if err1 != nil || err2 != nil {
			t.Fatalf("GetReferenceGasPrice: grpc=%v json=%v", err1, err2)
		}
		assertResponseEqual(t, "GetReferenceGasPrice", gr, jr)
	})

	t.Run("GetCurrentSystemState", func(t *testing.T) {
		gr, err1 := grpcClient.GetCurrentSystemState(ctx)
		jr, err2 := jsonClient.GetCurrentSystemState(ctx)
		if err1 != nil || err2 != nil {
			t.Fatalf("GetCurrentSystemState: grpc=%v json=%v", err1, err2)
		}
		assertResponseEqual(t, "GetCurrentSystemState", gr, jr)
	})

	t.Run("GetBalance", func(t *testing.T) {
		opts := GetBalanceOptions{Owner: testAddress}
		gr, err1 := grpcClient.GetBalance(ctx, opts)
		jr, err2 := jsonClient.GetBalance(ctx, opts)
		if err1 != nil || err2 != nil {
			t.Fatalf("GetBalance: grpc=%v json=%v", err1, err2)
		}
		assertResponseEqual(t, "GetBalance", gr, jr)
	})

	t.Run("ListBalances", func(t *testing.T) {
		opts := ListBalancesOptions{Owner: testAddress}
		gr, err1 := collectAllBalances(ctx, grpcClient, opts)
		jr, err2 := collectAllBalances(ctx, jsonClient, opts)
		if err1 != nil || err2 != nil {
			t.Fatalf("ListBalances: grpc=%v json=%v", err1, err2)
		}
		sort.Slice(gr.Balances, func(i, j int) bool { return gr.Balances[i].CoinType < gr.Balances[j].CoinType })
		sort.Slice(jr.Balances, func(i, j int) bool { return jr.Balances[i].CoinType < jr.Balances[j].CoinType })
		grCopy, jrCopy := *gr, *jr
		grCopy.Cursor, jrCopy.Cursor = nil, nil
		assertResponseEqual(t, "ListBalances", &grCopy, &jrCopy)
	})

	t.Run("ListCoins", func(t *testing.T) {
		opts := ListCoinsOptions{Owner: testAddress}
		gr, err1 := grpcClient.ListCoins(ctx, opts)
		jr, err2 := jsonClient.ListCoins(ctx, opts)
		if err1 != nil || err2 != nil {
			t.Fatalf("ListCoins: grpc=%v json=%v", err1, err2)
		}
		sort.Slice(gr.Objects, func(i, j int) bool { return gr.Objects[i].ObjectId < gr.Objects[j].ObjectId })
		sort.Slice(jr.Objects, func(i, j int) bool { return jr.Objects[i].ObjectId < jr.Objects[j].ObjectId })
		grCopy, jrCopy := *gr, *jr
		grCopy.Cursor, jrCopy.Cursor = nil, nil
		assertResponseEqual(t, "ListCoins", &grCopy, &jrCopy)
	})

	t.Run("GetCoinMetadata", func(t *testing.T) {
		opts := GetCoinMetadataOptions{CoinType: SUI_TYPE_ARG}
		gr, err1 := grpcClient.GetCoinMetadata(ctx, opts)
		jr, err2 := jsonClient.GetCoinMetadata(ctx, opts)
		if err1 != nil || err2 != nil {
			t.Fatalf("GetCoinMetadata: grpc=%v json=%v", err1, err2)
		}
		assertResponseEqual(t, "GetCoinMetadata", gr, jr)
	})

	t.Run("GetObjects", func(t *testing.T) {
		opts := GetObjectsOptions{
			ObjectIds: []string{testObjectId},
			Include: ObjectInclude{
				PreviousTransaction: true,
				JSON:                true,
			},
		}
		gr, err1 := grpcClient.GetObjects(ctx, opts)
		jr, err2 := jsonClient.GetObjects(ctx, opts)
		if err1 != nil || err2 != nil {
			t.Fatalf("GetObjects: grpc=%v json=%v", err1, err2)
		}
		assertResponseEqual(t, "GetObjects", gr, jr)
	})

	t.Run("ListOwnedObjects", func(t *testing.T) {
		opts := ListOwnedObjectsOptions{Owner: testAddress}
		gr, err1 := collectAllOwnedObjects(ctx, grpcClient, opts)
		jr, err2 := collectAllOwnedObjects(ctx, jsonClient, opts)
		if err1 != nil || err2 != nil {
			t.Fatalf("ListOwnedObjects: grpc=%v json=%v", err1, err2)
		}
		sort.Slice(gr.Objects, func(i, j int) bool { return gr.Objects[i].ObjectId < gr.Objects[j].ObjectId })
		sort.Slice(jr.Objects, func(i, j int) bool { return jr.Objects[i].ObjectId < jr.Objects[j].ObjectId })
		grCopy, jrCopy := *gr, *jr
		grCopy.Cursor, jrCopy.Cursor = nil, nil
		assertResponseEqualEx(t, "ListOwnedObjects", &grCopy, &jrCopy, nil, normalizeMoveTypes)
	})

	t.Run("GetTransaction", func(t *testing.T) {
		opts := GetTransactionOptions{
			Digest: testTxDigest,
			Include: TransactionInclude{
				Transaction:    true,
				Effects:        true,
				BalanceChanges: true,
				Events:         true,
				ObjectTypes:    true,
				Bcs:            true,
			},
		}
		gr, err1 := grpcClient.GetTransaction(ctx, opts)
		jr, err2 := jsonClient.GetTransaction(ctx, opts)
		if err1 != nil || err2 != nil {
			t.Fatalf("GetTransaction: grpc=%v json=%v", err1, err2)
		}
		assertResponseEqualEx(t, "GetTransaction", gr, jr, []string{"transaction.bcs"}, normalizeTxForCompare)
	})

	t.Run("ListDynamicFields", func(t *testing.T) {
		opts := ListDynamicFieldsOptions{ParentId: testParentId}
		gr, err1 := grpcClient.ListDynamicFields(ctx, opts)
		jr, err2 := jsonClient.ListDynamicFields(ctx, opts)
		if err1 != nil || err2 != nil {
			t.Fatalf("ListDynamicFields: grpc=%v json=%v", err1, err2)
		}
		// Ignore cursor
		grCopy, jrCopy := *gr, *jr
		grCopy.Cursor, jrCopy.Cursor = nil, nil
		assertResponseEqualEx(t, "ListDynamicFields", &grCopy, &jrCopy, nil, normalizeMoveTypes)
	})

	t.Run("GetMoveFunction", func(t *testing.T) {
		opts := GetMoveFunctionOptions{
			PackageId:  testPkgId,
			ModuleName: testModule,
			Name:       testFunction,
		}
		gr, err1 := grpcClient.GetMoveFunction(ctx, opts)
		jr, err2 := jsonClient.GetMoveFunction(ctx, opts)
		if err1 != nil || err2 != nil {
			t.Fatalf("GetMoveFunction: grpc=%v json=%v", err1, err2)
		}
		assertResponseEqual(t, "GetMoveFunction", gr, jr)
	})

	t.Run("DefaultNameServiceName", func(t *testing.T) {
		opts := DefaultNameServiceNameOptions{Address: testAddress}
		gr, err1 := grpcClient.DefaultNameServiceName(ctx, opts)
		jr, err2 := jsonClient.DefaultNameServiceName(ctx, opts)
		if err1 != nil || err2 != nil {
			t.Fatalf("DefaultNameServiceName: grpc=%v json=%v", err1, err2)
		}
		assertResponseEqual(t, "DefaultNameServiceName", gr, jr)
	})
}
