package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"strings"

	v2 "github.com/block-vision/sui-go-sdk/pb/sui/rpc/v2"
	"google.golang.org/protobuf/proto"
)

func main() {
	fmt.Println("=== Sui gRPC State Service Examples ===")

	fmt.Println("\n6. Testing ListOwnedObjects via gRPC-Web...")
	owner := "0x2701eb93b0ba04e1a9815381b6d2671893e9c43b90374049ee08fa1f0b8a3b3f"
	exampleListOwnedObjectsGrpcWeb(context.Background(), "https://edge-grpc-mainnet-endpoint.blockvision.org", "", owner)
}

// exampleListOwnedObjectsGrpcWeb implements ListOwnedObjects using gRPC-Web protocol
// gRPC-Web uses HTTP/1.1 POST requests with specific headers and binary encoding
func exampleListOwnedObjectsGrpcWeb(ctx context.Context, target string, token string, owner string) {
	fmt.Println("\n=== Testing ListOwnedObjects via gRPC-Web ===")

	// Convert gRPC endpoint to HTTP endpoint for gRPC-Web
	grpcWebURL := target
	if !strings.HasPrefix(grpcWebURL, "http") {
		// Remove port if present and add https://
		grpcWebURL = strings.Replace(grpcWebURL, ":443", "", 1)
		if !strings.HasPrefix(grpcWebURL, "http") {
			grpcWebURL = "https://" + grpcWebURL
		}
	}

	// gRPC-Web service path format: /<package>.<service>/<method>
	servicePath := "/sui.rpc.v2.StateService/ListOwnedObjects"
	fullURL := grpcWebURL + servicePath

	fmt.Printf("📡 gRPC-Web URL: %s\n", fullURL)
	fmt.Printf("👤 Owner: %s\n", owner)

	// Create the request message
	req := &v2.ListOwnedObjectsRequest{
		Owner: &owner,
	}
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		fmt.Printf("❌ Failed to marshal request: %v\n", err)
		return
	}

	// Encode request for gRPC-Web binary format
	// Format: [flags (1 byte)][message length (4 bytes, big-endian)][message bytes]
	flags := byte(0) // 0 = no compression, no trailer
	length := uint32(len(reqBytes))

	var encodedReq bytes.Buffer
	encodedReq.WriteByte(flags)
	binary.Write(&encodedReq, binary.BigEndian, length)
	encodedReq.Write(reqBytes)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", fullURL, &encodedReq)
	if err != nil {
		fmt.Printf("❌ Failed to create HTTP request: %v\n", err)
		return
	}

	// Set gRPC-Web headers
	httpReq.Header.Set("Content-Type", "application/grpc-web+proto")
	httpReq.Header.Set("Accept", "application/grpc-web+proto")
	httpReq.Header.Set("X-Grpc-Web", "1")
	httpReq.Header.Set("X-User-Agent", "grpc-web-go/1.0")

	// Add authentication headers (same as gRPC metadata)
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
		httpReq.Header.Set("X-Api-Key", token)
		httpReq.Header.Set("X-Token", token)
	}

	fmt.Printf("📤 Sending gRPC-Web request...\n")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("❌ gRPC-Web request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("📥 Response status: %d %s\n", resp.StatusCode, resp.Status)

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response: %v\n", err)
		return
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ gRPC-Web request failed with status %d\n", resp.StatusCode)
		fmt.Printf("   Response headers: %v\n", resp.Header)
		fmt.Printf("   Response body (first 200 bytes): %s\n", string(respBody[:min(200, len(respBody))]))
		return
	}

	// Check Content-Type
	contentType := resp.Header.Get("Content-Type")
	fmt.Printf("📋 Content-Type: %s\n", contentType)

	// Parse gRPC-Web response
	// Binary format: [flags (1 byte)][message length (4 bytes, big-endian)][message bytes]
	if len(respBody) < 5 {
		fmt.Printf("❌ Invalid gRPC-Web response format (too short: %d bytes)\n", len(respBody))
		fmt.Printf("   Response body: %v\n", respBody)
		return
	}

	flags = respBody[0]
	messageLength := binary.BigEndian.Uint32(respBody[1:5])

	fmt.Printf("📊 Response flags: %d, Message length: %d, Total body length: %d\n", flags, messageLength, len(respBody))

	if len(respBody) < int(5+messageLength) {
		fmt.Printf("❌ Response body shorter than expected (got %d, expected %d)\n", len(respBody), 5+messageLength)
		return
	}

	// Extract message bytes
	messageBytes := respBody[5 : 5+messageLength]

	// Unmarshal response
	listOwnedObjectsResp := &v2.ListOwnedObjectsResponse{}
	err = proto.Unmarshal(messageBytes, listOwnedObjectsResp)
	if err != nil {
		fmt.Printf("❌ Failed to unmarshal response: %v\n", err)
		fmt.Printf("   Message bytes (first 50): %v\n", messageBytes[:min(50, len(messageBytes))])
		return
	}

	// Print response
	fmt.Printf("\n✅ gRPC-Web ListOwnedObjects Response:\n")
	fmt.Printf("   Objects count: %d\n", len(listOwnedObjectsResp.GetObjects()))
	if len(listOwnedObjectsResp.GetObjects()) > 0 {
		fmt.Printf("   First object ID: %s\n", listOwnedObjectsResp.GetObjects()[0].GetObjectId())
		fmt.Printf("   First object version: %d\n", listOwnedObjectsResp.GetObjects()[0].GetVersion())
	}
	if len(listOwnedObjectsResp.GetNextPageToken()) > 0 {
		fmt.Printf("   Has next page: true (token length: %d bytes)\n", len(listOwnedObjectsResp.GetNextPageToken()))
	} else {
		fmt.Printf("   Has next page: false\n")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
