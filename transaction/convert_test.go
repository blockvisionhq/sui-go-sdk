package transaction

import (
	"encoding/hex"
	"testing"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/assert"
)

func TestSuiAddressConversion(t *testing.T) {
	// 32-byte hex address, with "0x" prefix
	original := models.SuiAddress("0x" + hex.EncodeToString(make([]byte, 32)))

	bytes, err := ConvertSuiAddressStringToBytes(original)
	assert.NoError(t, err)
	assert.NotNil(t, bytes)

	back := ConvertSuiAddressBytesToString(*bytes)
	assert.Equal(t, original, back)
}

func TestSuiAddressShortFormIsNormalized(t *testing.T) {
	shortAddress := models.SuiAddress("0x1234")
	bytes, err := ConvertSuiAddressStringToBytes(shortAddress)

	assert.NoError(t, err)
	assert.Equal(t, models.SuiAddress("0x0000000000000000000000000000000000000000000000000000000000001234"), ConvertSuiAddressBytesToString(*bytes))
}

func TestObjectDigestConversion(t *testing.T) {
	// 32-byte object digest encoded in base58
	raw := make([]byte, 32)
	encoded := base58.Encode(raw)
	original := models.ObjectDigest(encoded)

	bytes, err := ConvertObjectDigestStringToBytes(original)
	assert.NoError(t, err)
	assert.NotNil(t, bytes)

	back := ConvertObjectDigestBytesToString(*bytes)
	assert.Equal(t, original, back)
}

func TestObjectDigestInvalidLength(t *testing.T) {
	// Invalid base58-encoded digest (wrong length)
	invalid := models.ObjectDigest(base58.Encode([]byte("short")))
	_, err := ConvertObjectDigestStringToBytes(invalid)
	assert.Error(t, err)
}
