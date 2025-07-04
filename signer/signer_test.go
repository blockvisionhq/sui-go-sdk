package signer_test

import (
	"crypto/elliptic"
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/block-vision/sui-go-sdk/constant"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/stretchr/testify/assert"
)

var (
	testEd25519Key   = "suiprivkey1qz86u0u95ky0zfaenhqrge4483m8sh5h59fr0pkgnypncrqcnx8ws4safpm"
	testSecp256k1Key = "suiprivkey1qz3s8u9scv5fk6ma0wm8n2rsvmqwtykcer0tfwjeknt8xy96verxqrdxpcf"
	testSecp256r1Key = "suiprivkey1qqlfj0p5tqd6fshhvypswl068k8awt5y54vjmwx8zf07ct6829xqz89gjdp"
)

var (
	testEd25519Signature   = "2mRkjtvn7rYxIlRfNKXC0h0esH2HEAaihvpXFD2ReMUBghJjkTdi+bDL6/WT0reI3zEB2+IV+ywa+8xvqvzwAA=="
	testSecp256k1Signature = "n14lks5/kqxifoeucE2t8TiPTUogbCGCCFOOT4INz068SQaY+eHc3vqNG/s3AjGZFDApbsqYvymkBUx7An4KAA=="
	testSecp256r1Signature = "hhhBKSJP1JuC4DQP05WsTtGdcALQlll8BoBL2fsExjBh8qKyKjs9EqN5IXqBjtQpuv7V/eXBB6ytyG4TH7tjDA=="
)

var (
	testEd25519Pubkey   = "4c3f14681e53aab8321c67894d5dd0894846281e9eca2715e528b00fa572bb57"
	testSecp256k1Pubkey = "f0dace75124b87898830199178b90c208625559afb19956c9d004522cf4b3dd9"
	testSecp256r1Pubkey = "0fc82bba88b167627cefe26b085f1afe11c61a282f4e1eae3f575c7aae7fee05"
)

func TestDecodeSuiPrivateKey(t *testing.T) {
	// Test invalid prefix
	_, err := signer.DecodeSuiPrivateKey("wrongprefix1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq")
	assert.Error(t, err)

	// Test unknown schema flag (simulate with altered data)
	invalidKey := "suiprivkey1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzzzzzz"
	_, err = signer.DecodeSuiPrivateKey(invalidKey)
	assert.Error(t, err)
}

func TestValidatedSecpSignerConstructors(t *testing.T) {
	invalidScalars := [][]byte{
		make([]byte, 32),
		make([]byte, 31),
	}

	for _, secretKey := range invalidScalars {
		if got, err := signer.NewSecp256k1SignerFromSecretKey(secretKey); err == nil || got != nil {
			t.Fatal("NewSecp256k1SignerFromSecretKey() accepted an invalid private key")
		}
		if got, err := signer.NewSecp256r1SignerFromSecretKey(secretKey); err == nil || got != nil {
			t.Fatal("NewSecp256r1SignerFromSecretKey() accepted an invalid private key")
		}
	}
	if got, err := signer.NewSecp256k1SignerFromSecretKey(secp256k1.Params().N.Bytes()); err == nil || got != nil {
		t.Fatal("NewSecp256k1SignerFromSecretKey() accepted its curve order")
	}
	if got, err := signer.NewSecp256r1SignerFromSecretKey(elliptic.P256().Params().N.Bytes()); err == nil || got != nil {
		t.Fatal("NewSecp256r1SignerFromSecretKey() accepted its curve order")
	}

	if got := signer.NewSecp256k1Signer(make([]byte, 32)); got != nil {
		t.Fatal("legacy Secp256k1 constructor must return nil for an invalid private key")
	}
	if got := signer.NewSecp256r1Signer(make([]byte, 32)); got != nil {
		t.Fatal("legacy Secp256r1 constructor must return nil for an invalid private key")
	}
}

func TestKeypairSignaturesVerifyAcrossSchemes(t *testing.T) {
	tests := []struct {
		name    string
		keypair signer.Keypair
	}{
		{"Ed25519", signer.NewSigner(make([]byte, 32))},
		{"Secp256k1", signer.NewSecp256k1Signer(append(make([]byte, 31), 1))},
		{"Secp256r1", signer.NewSecp256r1Signer(append(make([]byte, 31), 1))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signed, err := tt.keypair.SignMessage(base64.StdEncoding.EncodeToString([]byte("compatibility test")), constant.PersonalMessageIntentScope)
			if err != nil {
				t.Fatalf("SignPersonalMessage() error = %v", err)
			}

			parsed, err := models.FromSerializedSignature(signed.Signature)
			if err != nil {
				t.Fatalf("FromSerializedSignature() error = %v", err)
			}
			if got, want := len(parsed.PubKey), len(tt.keypair.PublicKeyBytes()); got != want {
				t.Fatalf("public key length = %d, want %d", got, want)
			}

			gotSigner, verified, err := models.VerifyPersonalMessage("compatibility test", signed.Signature)
			if err != nil {
				t.Fatalf("VerifyPersonalMessage() error = %v", err)
			}
			if !verified {
				t.Fatal("VerifyPersonalMessage() did not verify signature")
			}
			if gotSigner != tt.keypair.GetAddress() {
				t.Fatalf("signer = %s, want %s", gotSigner, tt.keypair.GetAddress())
			}
		})
	}
}

func TestSignerFromSuiSecret(t *testing.T) {
	tests := []struct {
		name           string
		encoded        string
		expectedSig    string
		expectedPubkey string
	}{
		{"Ed25519", testEd25519Key, testEd25519Signature, testEd25519Pubkey},
		{"Secp256k1", testSecp256k1Key, testSecp256k1Signature, testSecp256k1Pubkey},
		{"Secp256r1", testSecp256r1Key, testSecp256r1Signature, testSecp256r1Pubkey},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signer, err := signer.SignerFromSuiSecret(tt.encoded)
			assert.NoError(t, err)
			assert.NotNil(t, signer)

			pubKeyBytes := signer.PublicKeyBytes()
			pubKeyHex := hex.EncodeToString(pubKeyBytes)
			assert.NotEmpty(t, pubKeyHex)

			msg := []byte("test message")
			sig, err := signer.Sign(msg)

			assert.Equal(t, tt.expectedSig, base64.StdEncoding.EncodeToString(sig))
			assert.NoError(t, err)
			assert.NotEmpty(t, sig)
			assert.NotEmpty(t, pubKeyBytes)
			assert.Equal(t, tt.expectedPubkey, pubKeyHex)
		})
	}
}
