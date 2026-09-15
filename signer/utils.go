package signer

import (
	"encoding/hex"

	"golang.org/x/crypto/blake2b"
)

func toSuiAddress(pubKey []byte, flag byte) string {
	tmp := []byte{flag}
	tmp = append(tmp, pubKey...)
	addrBytes := blake2b.Sum256(tmp)
	addr := "0x" + hex.EncodeToString(addrBytes[:])[:AddressLength]
	return addr
}
