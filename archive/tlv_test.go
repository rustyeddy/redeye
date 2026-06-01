package redeye

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var Config = struct {
	Debug bool
}{}

func TestTLV(t *testing.T) {
	const tlvLength = 4
	tlv := NewTLV(CMDPlay, 4)
	require.Equal(t, CMDPlay, tlv.Type())
	require.Equal(t, tlvLength, tlv.Len())

	ty, l := tlv.TypeLen()
	assert.Equal(t, int(CMDPlay), ty)
	assert.Equal(t, tlvLength, l)
	assert.Equal(t, tlvLength-2, len(tlv.Value()))
	assert.Equal(t, string([]byte{CMDPlay, tlvLength, 0, 0}), tlv.Str())
}

func TestTLVLenHandlesNil(t *testing.T) {
	var nilTLV *TLV
	assert.Equal(t, 0, nilTLV.Len())
	assert.Equal(t, 0, (&TLV{}).Len())
}

func TestNewTLVPanicsOnShortLength(t *testing.T) {
	assert.Panics(t, func() {
		NewTLV(CMDPlay, 1)
	})
}
