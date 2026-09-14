package transaction

import (
	"encoding/base64"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/block-vision/sui-go-sdk/mystenbcs"
)

func TestBcsUnmarshal(t *testing.T) {
	b64Tx := "AAAIAQBiBnaLHr5TzVJu29SyLYueU6dFUyLeaY9ZeK11c+8IrXPUayEAAAAAIP3zvI2RTxZdaMnH47J4OJQbyHdIb50BtZ2JxNb4gD7AAQDo6vegaubVlLDf0hxZfOtmAkBc4ryrHxXfLLjFo72CF4bUayEAAAAAIECXRV/tLtwVmiAi/xXB6L5v1KFuEZQ0oah1NwYdBcLfAAgAhNcXAAAAAAEB2qRikmMsPE2PMfI+oPmzaij/NnfpaEmA5EOEA6Z6PY8uBRgAAAAAAAABAa2qRWjdYdo2MMQNLRRttp5jmUbSJZnqQxXT0+oBmR8aY8ddIQAAAAABAQFjm15DPaMXOegAzQhfNW5kyuIilm0PGxG9ncdrMi/1i2LvCxMAAAAAAQEBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAYBAAAAAAAAAAAACE7aKJKcAwAABwMBAAABAQEAAgEAAAEBAgACAwEAAAABAQIAADhkx8WaSIn+wF0arkvJ26Wg4JQFlLQk++1Eyz9qxMAyBWNldHVzCHN3YXBfYjJhAgfhtFoOZBuZVaIKoK0cH0rYaq2K+wcpbUCF40mlDpC9ygRibHVlBEJMVUUAB9ujRnLjDLBlsfk+OrVTGHaP1v72bBWULJ98uEbi+QDnBHVzZGMEVVNEQwAFAQMAAQQAAQUAAwIAAAABBgAAJIX+udQsfDvLjs3lVa1A8bBz2ftPrzVPotMKCxg6I84FdXRpbHMYdHJhbnNmZXJfb3JfZGVzdHJveV9jb2luAQfbo0Zy4wywZbH5Pjq1Uxh2j9b+9mwVlCyffLhG4vkA5wR1c2RjBFVTREMAAQMBAAAAACSF/rnULHw7y47N5VWtQPGwc9n7T681T6LTCgsYOiPOBXV0aWxzFGNoZWNrX2NvaW5fdGhyZXNob2xkAQfhtFoOZBuZVaIKoK0cH0rYaq2K+wcpbUCF40mlDpC9ygRibHVlBEJMVUUAAgMDAAAAAQcAACSF/rnULHw7y47N5VWtQPGwc9n7T681T6LTCgsYOiPOBXV0aWxzGHRyYW5zZmVyX29yX2Rlc3Ryb3lfY29pbgEH4bRaDmQbmVWiCqCtHB9K2GqtivsHKW1AheNJpQ6QvcoEYmx1ZQRCTFVFAAEDAwAAAPGLdfcB6yxirRUeIzQYdrBeeCZsC49tvfE1MjQ3HRe+AeFyymxHLoC1zuCDjWceFIVGpQinWoTY3E/nQtJ0PXeLhtRrIQAAAAAg7wnlIFKxbd6S0eZ7pktb3VGeUCezpe3rCIK/6e1qR+7xi3X3AessYq0VHiM0GHawXngmbAuPbb3xNTI0Nx0Xvu4CAAAAAAAAAOH1BQAAAAAA"
	txBz, _ := base64.StdEncoding.DecodeString(b64Tx)

	var txData TransactionData
	_, err := mystenbcs.Unmarshal(txBz, &txData)
	require.NoError(t, err)

	genBz, err := txData.Marshal()
	require.NoError(t, err)
	require.Equal(t, txBz, genBz)
}

// TestMakeMoveVecBcs checks the Option tag byte of MakeMoveVec.Type. A vector of
// objects carries no type and has to encode the None tag, otherwise every byte
// after it is shifted.
func TestMakeMoveVecBcs(t *testing.T) {
	cases := []struct {
		name    string
		command Command
		want    []byte
	}{
		{
			name: "no type",
			command: Command{MakeMoveVec: &MakeMoveVec{
				Type:     nil,
				Elements: []*Argument{{Input: lo.ToPtr(uint16(0))}, {Input: lo.ToPtr(uint16(1))}},
			}},
			want: []byte{0x05, 0x00, 0x02, 0x01, 0x00, 0x00, 0x01, 0x01, 0x00},
		},
		{
			name: "with type",
			command: Command{MakeMoveVec: &MakeMoveVec{
				Type:     lo.ToPtr("u64"),
				Elements: []*Argument{{Input: lo.ToPtr(uint16(0))}},
			}},
			want: []byte{0x05, 0x01, 0x03, 0x75, 0x36, 0x34, 0x01, 0x01, 0x00, 0x00},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bz, err := mystenbcs.Marshal(&c.command)
			require.NoError(t, err)
			require.Equal(t, c.want, bz)
		})
	}
}

// TestArgumentBcs pins the encoding of the Argument enum. Its variants are
// pointers too, and they must keep carrying only the variant index.
func TestArgumentBcs(t *testing.T) {
	cases := []struct {
		name     string
		argument Argument
		want     []byte
	}{
		{
			name:     "gas coin",
			argument: Argument{GasCoin: struct{}{}},
			want:     []byte{0x00},
		},
		{
			name:     "input",
			argument: Argument{Input: lo.ToPtr(uint16(3))},
			want:     []byte{0x01, 0x03, 0x00},
		},
		{
			name:     "result",
			argument: Argument{Result: lo.ToPtr(uint16(3))},
			want:     []byte{0x02, 0x03, 0x00},
		},
		{
			name:     "nested result",
			argument: Argument{NestedResult: &NestedResult{Index: 1, ResultIndex: 2}},
			want:     []byte{0x03, 0x01, 0x00, 0x02, 0x00},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bz, err := mystenbcs.Marshal(&c.argument)
			require.NoError(t, err)
			require.Equal(t, c.want, bz)
		})
	}
}
