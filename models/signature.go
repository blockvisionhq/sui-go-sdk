package models

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"

	"golang.org/x/crypto/blake2b"

	"github.com/block-vision/sui-go-sdk/constant"
	"github.com/block-vision/sui-go-sdk/mystenbcs"
	"github.com/block-vision/sui-go-sdk/zklogin"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	secp256k1ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

type InputObjectKind map[string]interface{}
type ObjectId = HexData
type Digest = Base64Data

type ObjectRef struct {
	Digest   string   `json:"digest"`
	ObjectId ObjectId `json:"objectId"`
	Version  uint64   `json:"version"`
}

type SigScheme string

const (
	SigEd25519   SigScheme = "ED25519"
	SigSecp256k1 SigScheme = "Secp256k1"
)

type SigFlag byte

const (
	SigFlagEd25519   SigFlag = 0x00
	SigFlagSecp256k1 SigFlag = 0x01
	SigFlagSecp256r1 SigFlag = 0x02
)

type HexData struct {
	data []byte
}

type SignaturePubkeyPair struct {
	SignatureScheme string
	Signature       []byte
	PubKey          []byte
}

func NewHexData(str string) (*HexData, error) {
	if strings.HasPrefix(str, "0x") || strings.HasPrefix(str, "0X") {
		str = str[2:]
	}
	data, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}
	return &HexData{data}, nil
}

func (a HexData) Data() []byte {
	return a.data
}

type Bytes []byte

func (b Bytes) GetHexData() HexData {
	return HexData{b}
}
func (b Bytes) GetBase64Data() Base64Data {
	return Base64Data{b}
}

type Base64Data struct {
	data []byte
}

func NewBase64Data(str string) (*Base64Data, error) {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}
	return &Base64Data{data}, nil
}

func (h Base64Data) Data() []byte {
	return h.data
}

type SignedTransaction struct {
	// transaction data bytes
	TxBytes string `json:"tx_bytes"`

	// Flag of the signature scheme that is used.
	SigScheme SigScheme `json:"sig_scheme"`

	// transaction signature
	Signature *Base64Data `json:"signature"`

	// signer's public key
	PublicKey *Base64Data `json:"pub_key"`
}

type SignedTransactionSerializedSig struct {
	// transaction data bytes
	TxBytes string `json:"tx_bytes"`

	// transaction signature
	Signature string `json:"signature"`
}

var IntentBytes = []byte{0, 0, 0}

func (txn *TxnMetaData) SignSerializedSigWith(privateKey ed25519.PrivateKey) *SignedTransactionSerializedSig {
	txBytes, _ := base64.StdEncoding.DecodeString(txn.TxBytes)
	message := messageWithIntent(txBytes)
	digest := blake2b.Sum256(message)
	var noHash crypto.Hash
	sigBytes, err := privateKey.Sign(nil, digest[:], noHash)
	if err != nil {
		log.Fatal(err)
	}
	return &SignedTransactionSerializedSig{
		TxBytes:   txn.TxBytes,
		Signature: ToSerializedSignature(sigBytes, privateKey.Public().(ed25519.PublicKey)),
	}
}

func messageWithIntent(message []byte) []byte {
	intent := IntentBytes
	intentMessage := make([]byte, len(intent)+len(message))
	copy(intentMessage, intent)
	copy(intentMessage[len(intent):], message)
	return intentMessage
}

// ToSerializedSignature serializes an Ed25519 signature. It is retained for
// compatibility with existing callers.
func ToSerializedSignature(signature, pubKey []byte) string {
	return ToSerializedSignatureWithScheme(signature, pubKey, byte(SigFlagEd25519))
}

// ToSerializedSignatureWithScheme serializes a signature using the supplied
// Sui signature-scheme flag.
func ToSerializedSignatureWithScheme(signature, pubKey []byte, sigFlag byte) string {
	signatureLen := len(signature)
	pubKeyLen := len(pubKey)
	serializedSignature := make([]byte, 1+signatureLen+pubKeyLen)
	serializedSignature[0] = byte(sigFlag)
	copy(serializedSignature[1:], signature)
	copy(serializedSignature[1+signatureLen:], pubKey)
	return base64.StdEncoding.EncodeToString(serializedSignature)
}

func FromSerializedSignature(serializedSignature string) (*SignaturePubkeyPair, error) {
	if strings.EqualFold(serializedSignature, "") {
		return nil, fmt.Errorf("multiSig is not supported")
	}

	_bytes, err := base64.StdEncoding.DecodeString(serializedSignature)
	if err != nil {
		return nil, err
	}
	if len(_bytes) == 0 {
		return nil, fmt.Errorf("serialized signature is empty")
	}

	if _bytes[0] == 3 {
		// A multisig payload contains multiple public keys and signatures, so it
		// does not have a single fixed-width public-key suffix.
		return &SignaturePubkeyPair{
			SignatureScheme: "MultiSig",
			Signature:       _bytes[1:],
		}, nil
	}
	if _bytes[0] == 5 {
		parsed, err := zklogin.ParseSerializedZkLoginSignature(serializedSignature)
		if err != nil {
			return nil, err
		}
		return &SignaturePubkeyPair{
			SignatureScheme: string(parsed.SignatureScheme),
			Signature:       parsed.Signature,
			PubKey:          parsed.PubKey,
		}, nil
	}

	signatureScheme, publicKeyLength, err := signatureSchemeAndPublicKeyLength(_bytes[0])
	if err != nil {
		return nil, err
	}
	if len(_bytes) <= 1+publicKeyLength {
		return nil, fmt.Errorf("serialized signature is too short for %s", signatureScheme)
	}

	signature := _bytes[1 : len(_bytes)-publicKeyLength]
	pubKeyBytes := _bytes[1+len(signature):]

	keyPair := &SignaturePubkeyPair{
		SignatureScheme: signatureScheme,
		Signature:       signature,
		PubKey:          pubKeyBytes,
	}
	return keyPair, nil
}

func signatureSchemeAndPublicKeyLength(scheme byte) (string, int, error) {
	switch scheme {
	case 0:
		return "ED25519", ed25519.PublicKeySize, nil
	case 1:
		return "Secp256k1", 33, nil
	case 2:
		return "Secp256r1", 33, nil
	default:
		return "", 0, fmt.Errorf("signature flag %d is not supported", scheme)
	}
}

func VerifyPersonalMessage(message string, signature string) (signer string, pass bool, err error) {
	b64Message := base64.StdEncoding.EncodeToString([]byte(message))
	return VerifyMessage(b64Message, signature, constant.PersonalMessageIntentScope)
}

func VerifyTransaction(b64Message string, signature string) (signer string, pass bool, err error) {
	return VerifyMessage(b64Message, signature, constant.TransactionDataIntentScope)
}

func VerifyMessage(message, signature string, scope constant.IntentScope) (signer string, pass bool, err error) {
	b64Bytes, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		return "", false, err
	}

	bcsEncodedMsg := bytes.Buffer{}
	bcsEncoder := mystenbcs.NewEncoder(&bcsEncodedMsg)
	if err := bcsEncoder.Encode(b64Bytes); err != nil {
		return "", false, err
	}

	serializedSignature, err := FromSerializedSignature(signature)
	if err != nil {
		return "", false, err
	}

	for _, payload := range [][]byte{b64Bytes, bcsEncodedMsg.Bytes()} {
		digest := blake2b.Sum256(NewMessageWithIntent(payload, scope))
		pass, err = verifyDigest(serializedSignature, digest[:])
		if err != nil {
			return "", false, err
		}
		if pass {
			break
		}
	}

	signer = PublicKeyToSuiAddress(serializedSignature.PubKey, signatureSchemeFlag(serializedSignature.SignatureScheme))

	return
}

func signatureSchemeFlag(signatureScheme string) byte {
	switch signatureScheme {
	case "Secp256k1":
		return byte(SigFlagSecp256k1)
	case "Secp256r1":
		return byte(SigFlagSecp256r1)
	default:
		return byte(SigFlagEd25519)
	}
}

func verifyDigest(serializedSignature *SignaturePubkeyPair, digest []byte) (bool, error) {
	switch serializedSignature.SignatureScheme {
	case "ED25519":
		return ed25519.Verify(serializedSignature.PubKey, digest, serializedSignature.Signature), nil
	case "Secp256k1":
		if len(serializedSignature.Signature) != 64 {
			return false, fmt.Errorf("invalid Secp256k1 signature length: %d", len(serializedSignature.Signature))
		}
		publicKey, err := secp256k1.ParsePubKey(serializedSignature.PubKey)
		if err != nil {
			return false, err
		}
		var r, s secp256k1.ModNScalar
		if r.SetByteSlice(serializedSignature.Signature[:32]) || s.SetByteSlice(serializedSignature.Signature[32:]) {
			return false, fmt.Errorf("invalid Secp256k1 signature scalar")
		}
		hash := sha256.Sum256(digest)
		return secp256k1ecdsa.NewSignature(&r, &s).Verify(hash[:], publicKey), nil
	case "Secp256r1":
		if len(serializedSignature.Signature) != 64 {
			return false, fmt.Errorf("invalid Secp256r1 signature length: %d", len(serializedSignature.Signature))
		}
		x, y := elliptic.UnmarshalCompressed(elliptic.P256(), serializedSignature.PubKey)
		if x == nil || y == nil {
			return false, fmt.Errorf("invalid Secp256r1 public key")
		}
		hash := sha256.Sum256(digest)
		return ecdsa.Verify(&ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, hash[:], new(big.Int).SetBytes(serializedSignature.Signature[:32]), new(big.Int).SetBytes(serializedSignature.Signature[32:])), nil
	default:
		return false, fmt.Errorf("signature scheme %s is not supported", serializedSignature.SignatureScheme)
	}
}

func Ed25519PublicKeyToSuiAddress(pubKey []byte) string {
	return PublicKeyToSuiAddress(pubKey, byte(SigFlagEd25519))
}

// PublicKeyToSuiAddress derives a Sui address for the supplied public key and
// signature-scheme flag.
func PublicKeyToSuiAddress(pubKey []byte, sigFlag byte) string {
	newPubkey := []byte{sigFlag}
	newPubkey = append(newPubkey, pubKey...)

	addrBytes := blake2b.Sum256(newPubkey)
	return fmt.Sprintf("0x%s", hex.EncodeToString(addrBytes[:])[:64])
}
