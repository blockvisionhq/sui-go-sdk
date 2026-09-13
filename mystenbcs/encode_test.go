package mystenbcs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type optionalPayload struct {
	Opt *string `bcs:"optional"`
	N   uint8
}

type requiredPayload struct {
	Ptr *uint8
}

// TestMarshalOptional checks the Option tag byte is written for both variants:
// 0x00 for None, 0x01 followed by the encoded value for Some.
func TestMarshalOptional(t *testing.T) {
	cases := []struct {
		name  string
		value optionalPayload
		want  []byte
	}{
		{
			name:  "none",
			value: optionalPayload{Opt: nil, N: 7},
			want:  []byte{0x00, 0x07},
		},
		{
			name:  "some",
			value: optionalPayload{Opt: ptr("x"), N: 7},
			want:  []byte{0x01, 0x01, 0x78, 0x07},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bz, err := Marshal(c.value)
			require.NoError(t, err)
			require.Equal(t, c.want, bz)

			var got optionalPayload
			n, err := Unmarshal(bz, &got)
			require.NoError(t, err)
			require.Equal(t, len(c.want), n)
			require.Equal(t, c.value, got)
		})
	}
}

// TestMarshalNilPointer checks a nil pointer that is not marked optional is
// rejected rather than encoded as no bytes at all.
func TestMarshalNilPointer(t *testing.T) {
	_, err := Marshal(requiredPayload{})
	require.Error(t, err)
}

func ptr[T any](v T) *T {
	return &v
}
