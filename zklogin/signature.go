package zklogin

import (
	"fmt"

	"github.com/fardream/go-bcs/bcs"

	"github.com/block-vision/sui-go-sdk/mystenbcs"
)

func parseZkLoginSignature(signature interface{}) (*ZkLoginSignature, error) {
	var bytes []byte
	var err error

	// Check if the input is a base64 string or a byte array
	switch sig := signature.(type) {
	case string:
		bytes, err = mystenbcs.FromBase64(sig)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64: %v", err)
		}
	case []byte:
		bytes = sig
	default:
		return nil, fmt.Errorf("unsupported input type")
	}

	// Deserialize the bytes into ZkLoginSignature struct using BCS
	var zkSig ZkLoginSignature
	_, err = bcs.Unmarshal(bytes, &zkSig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse BCS data: %v", err)
	}

	return &zkSig, nil
}
