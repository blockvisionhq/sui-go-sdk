package signer

import (
	"bytes"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/block-vision/sui-go-sdk/common/keypair"
	"github.com/block-vision/sui-go-sdk/constant"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/mystenbcs"
	"github.com/btcsuite/btcutil/bech32"
	"github.com/cosmos/go-bip39"
	"golang.org/x/crypto/blake2b"
)

const (
	SigntureFlagEd25519     = 0x0
	SigntureFlagSecp256k1   = 0x1
	AddressLength           = 64
	DerivationPathEd25519   = `m/44'/784'/0'/0'/0'`
	DerivationPathSecp256k1 = `m/54'/784'/0'/0/0`
)

const SigFlagEd25519 = 0x00

type Ed25519Signer Signer

type Signer struct {
	PriKey  ed25519.PrivateKey
	PubKey  ed25519.PublicKey
	Address string
}

func NewSigner(seed []byte) *Signer {
	priKey := ed25519.NewKeyFromSeed(seed[:])
	pubKey := priKey.Public().(ed25519.PublicKey)

	tmp := []byte{byte(keypair.Ed25519Flag)}
	tmp = append(tmp, pubKey...)
	addrBytes := blake2b.Sum256(tmp)
	addr := "0x" + hex.EncodeToString(addrBytes[:])[:AddressLength]

	return &Signer{
		PriKey:  priKey,
		PubKey:  pubKey,
		Address: addr,
	}
}

func NewSignerWithSecretKey(secret string) (*Signer, error) {
	hrp, data, err := bech32.Decode(secret)
	if err != nil {
		return nil, err
	}
	if hrp != "suiprivkey" {
		return nil, fmt.Errorf("Invalid bech32 prefix: %s", hrp)
	}

	// bech32 5bit to 8bit
	decoded, err := bech32.ConvertBits(data, 5, 8, false)
	if err != nil {
		return nil, err
	}
	if len(decoded) < 2 {
		return nil, fmt.Errorf("Invalid bech32 data length: %d", len(decoded))
	}

	privKey := decoded[1:]
	return NewSigner(privKey), nil
}

func NewSignertWithMnemonic(mnemonic string) (*Signer, error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return nil, err
	}
	key, err := DeriveForPath("m/44'/784'/0'/0'/0'", seed)
	if err != nil {
		return nil, err
	}
	return NewSigner(key.Key), nil
}

func (s *Signer) GetAddress() string {
	return s.Address
}

func (s *Signer) Schema() byte {
	return byte(0x0)
}

func (s *Signer) PublicKeyBytes() []byte {
	return s.PubKey
}

func (s *Signer) SignMessage(data string, scope constant.IntentScope) (*SignedMessageSerializedSig, error) {
	txBytes, _ := base64.StdEncoding.DecodeString(data)
	message := models.NewMessageWithIntent(txBytes, scope)
	digest := blake2b.Sum256(message)
	var noHash crypto.Hash
	sigBytes, err := s.PriKey.Sign(nil, digest[:], noHash)
	if err != nil {
		return nil, err
	}

	ret := &SignedMessageSerializedSig{
		Message:   data,
		Signature: models.ToSerializedSignature(sigBytes, s.PriKey.Public().(ed25519.PublicKey)),
	}
	return ret, nil
}

func (s *Signer) Sign(message []byte) ([]byte, error) {
	digest := blake2b.Sum256(message)
	var noHash crypto.Hash
	sig, err := s.PriKey.Sign(rand.Reader, digest[:], noHash)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func (s *Signer) SignTransaction(b64TxBytes string) (*models.SignedTransactionSerializedSig, error) {
	result, err := s.SignMessage(b64TxBytes, constant.PersonalMessageIntentScope)
	if err != nil {
		return nil, err
	}

	return &models.SignedTransactionSerializedSig{
		TxBytes:   result.Message,
		Signature: result.Signature,
	}, nil
}

func (s *Signer) SignPersonalMessage(message string) (*SignedMessageSerializedSig, error) {
	b64Message := base64.StdEncoding.EncodeToString([]byte(message))
	return s.SignMessage(b64Message, constant.PersonalMessageIntentScope)
}

// SignPersonalMessageV1 is the same as SignPersonalMessage, but it uses the new message format for personal messages.
func (s *Signer) SignPersonalMessageV1(message string) (*SignedMessageSerializedSig, error) {
	b64Message := base64.StdEncoding.EncodeToString([]byte(message))
	return s.SignMessageV1(b64Message, constant.PersonalMessageIntentScope)
}

// SignMessageV1 is the same as SignMessage, but it uses the new message format for personal messages.
func (s *Signer) SignMessageV1(data string, scope constant.IntentScope) (*SignedMessageSerializedSig, error) {
	b64Bytes, _ := base64.StdEncoding.DecodeString(data)

	bcsEncodedMsg := bytes.Buffer{}
	bcsEncoder := mystenbcs.NewEncoder(&bcsEncodedMsg)
	bcsEncoder.Encode(b64Bytes)
	message := models.NewMessageWithIntent(bcsEncodedMsg.Bytes(), scope)
	digest := blake2b.Sum256(message)
	var noHash crypto.Hash
	sigBytes, err := s.PriKey.Sign(nil, digest[:], noHash)
	if err != nil {
		return nil, err
	}

	ret := &SignedMessageSerializedSig{
		Message:   data,
		Signature: models.ToSerializedSignature(sigBytes, s.PriKey.Public().(ed25519.PublicKey)),
	}
	return ret, nil
}
