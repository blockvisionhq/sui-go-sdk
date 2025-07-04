package signer

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"math/big"

	"github.com/block-vision/sui-go-sdk/constant"
	"github.com/block-vision/sui-go-sdk/models"
	"golang.org/x/crypto/blake2b"
)

const SigFlagSecp256r1 = 0x02

type Secp256r1Signer struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	SuiAddress string
}

// NewSecp256r1Signer creates a Secp256r1 signer for compatibility with
// existing callers. It returns nil when secretKey is invalid; new callers
// should use NewSecp256r1SignerFromSecretKey to receive the validation error.
func NewSecp256r1Signer(secretKey []byte) *Secp256r1Signer {
	signer, _ := NewSecp256r1SignerFromSecretKey(secretKey)
	return signer
}

// NewSecp256r1SignerFromSecretKey creates a Secp256r1 signer from a valid
// 32-byte private scalar.
func NewSecp256r1SignerFromSecretKey(secretKey []byte) (*Secp256r1Signer, error) {
	curve := elliptic.P256()
	if err := validateSecretScalar(secretKey, curve.Params().N, "secp256r1"); err != nil {
		return nil, err
	}
	d := new(big.Int).SetBytes(secretKey)

	priv := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
		},
		D: d,
	}
	priv.X, priv.Y = curve.ScalarBaseMult(secretKey)

	pubBytes := elliptic.MarshalCompressed(curve, priv.X, priv.Y)
	addr := toSuiAddress(pubBytes, SigFlagSecp256r1)

	return &Secp256r1Signer{
		PrivateKey: priv,
		PublicKey:  &priv.PublicKey,
		SuiAddress: addr,
	}, nil

}

func (s *Secp256r1Signer) Sign(message []byte) ([]byte, error) {
	digest := blake2b.Sum256(message)
	msgHash := sha256.Sum256(digest[:])

	r, ss, err := ecdsa.Sign(rand.Reader, s.PrivateKey, msgHash[:])
	if err != nil {
		return nil, err
	}

	rBytes := r.Bytes()
	sBytes := ss.Bytes()

	rPadded := make([]byte, 32)
	sPadded := make([]byte, 32)

	copy(rPadded[32-len(rBytes):], rBytes)
	copy(sPadded[32-len(sBytes):], sBytes)

	rawSig := append(rPadded, sPadded...)

	return rawSig, nil
}

func (s *Secp256r1Signer) SignMessage(data string, scope constant.IntentScope) (*SignedMessageSerializedSig, error) {
	txBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	message := models.NewMessageWithIntent(txBytes, scope)
	digest := blake2b.Sum256(message)
	hash := sha256.Sum256(digest[:])

	r, ss, err := ecdsa.Sign(rand.Reader, s.PrivateKey, hash[:])
	if err != nil {
		return nil, err
	}

	rBytes := r.Bytes()
	sBytes := ss.Bytes()

	rPadded := make([]byte, 32)
	sPadded := make([]byte, 32)
	copy(rPadded[32-len(rBytes):], rBytes)
	copy(sPadded[32-len(sBytes):], sBytes)

	rawSig := append(rPadded, sPadded...)
	pubBytes := elliptic.MarshalCompressed(s.PublicKey.Curve, s.PublicKey.X, s.PublicKey.Y)

	return &SignedMessageSerializedSig{
		Message:   data,
		Signature: models.ToSerializedSignatureWithScheme(rawSig, pubBytes, SigFlagSecp256r1),
	}, nil

}

func (s *Secp256r1Signer) GetAddress() string {
	return s.SuiAddress
}

func (s *Secp256r1Signer) PublicKeyBytes() []byte {
	return elliptic.MarshalCompressed(s.PrivateKey.Curve, s.PrivateKey.X, s.PrivateKey.Y)
}

func (s *Secp256r1Signer) Schema() byte {
	return byte(SigFlagSecp256r1)
}
