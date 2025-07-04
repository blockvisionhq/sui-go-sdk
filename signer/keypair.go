package signer

import (
	"github.com/block-vision/sui-go-sdk/constant"
)

type SignedMessageSerializedSig struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

type Keypair interface {
	Sign(message []byte) ([]byte, error)
	SignMessage(data string, scope constant.IntentScope) (*SignedMessageSerializedSig, error)

	Schema() byte
	GetAddress() string
	PublicKeyBytes() []byte
}
