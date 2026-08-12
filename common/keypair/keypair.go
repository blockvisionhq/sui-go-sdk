package keypair

import (
	"encoding/base64"
	"github.com/block-vision/sui-go-sdk/common/sui_error"
	"math"
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/ed25519"
)

type KeyPair byte

const (
	Ed25519Flag   KeyPair = 0
	Secp256k1Flag KeyPair = 1
	ErrorFlag     byte    = math.MaxUint8
)

const (
	ed25519PublicKeyLength   = 32
	secp256k1PublicKeyLength = 33
)

const (
	DefaultAccountAddressLength = 16
	AccountAddress20Length      = 20
	AccountAddress32Length      = 32
)

func FetchKeyPair(value string) (models.SuiKeyPair, error) {
	result, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return models.SuiKeyPair{}, err
	}
	if len(result) == 0 {
		return models.SuiKeyPair{}, err
	}
	switch result[0] {
	case byte(Ed25519Flag):
		pb := result[1 : ed25519PublicKeyLength+1]
		sk := result[1+ed25519PublicKeyLength:]
		pbInBase64 := encodeBase64(pb)
		return models.SuiKeyPair{
			Flag:            byte(Ed25519Flag),
			PrivateKey:      sk,
			PublicKeyBase64: pbInBase64,
			PublicKey:       pb,
			Address:         fromPublicKeyBytesToAddress(pb, byte(Ed25519Flag)),
		}, nil
	case byte(Secp256k1Flag):
		pb := result[1 : secp256k1PublicKeyLength+1]
		sk := result[1+secp256k1PublicKeyLength:]
		pbInBase64 := encodeBase64(pb)
		return models.SuiKeyPair{
			Flag:            byte(Secp256k1Flag),
			PrivateKey:      sk,
			PublicKey:       pb,
			PublicKeyBase64: pbInBase64,
			Address:         fromPublicKeyBytesToAddress(pb, byte(Secp256k1Flag)),
		}, nil
	default:
		return models.SuiKeyPair{}, sui_error.ErrInvalidEncryptFlag
	}
}

func CreateNewKeyPair() {
	// 1. Generate 24-word mnemonic
	entropy, _ := bip39.NewEntropy(256)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	fmt.Println("Mnemonic:", mnemonic)

	// 2. Convert mnemonic to seed
	seed := bip39.NewSeed(mnemonic, "")

	// 3. SLIP-0010 Ed25519 master key
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		log.Fatal(err)
	}

	// 4. Sui derivation path: m/44'/784'/0'/0'/0'
	path := []uint32{
		44 + bip32.FirstHardenedChild,
		784 + bip32.FirstHardenedChild,
		0 + bip32.FirstHardenedChild,
		0 + bip32.FirstHardenedChild,
		0 + bip32.FirstHardenedChild,
	}

	key := masterKey
	for _, p := range path {
		key, err = key.NewChildKey(p)
		if err != nil {
			log.Fatal(err)
		}
	}

	// 5. Private key seed (32 bytes)
	seed32 := key.Key

	// 6. Generate Ed25519 keypair (correct Go version)
	pub, priv, err := ed25519.GenerateKey(bytes.NewReader(seed32))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Private Key:", hex.EncodeToString(priv))
	fmt.Println("Public Key :", hex.EncodeToString(pub))

	// 7. Convert public key to Sui address
	hash := blake2b.Sum256(pub)
	address := "0x" + hex.EncodeToString(hash[:])
	fmt.Println("Sui Address:", address)

}
