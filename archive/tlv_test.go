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
	tlv := NewTLV(CMDPlay, 4)
	require.Equal(t, CMDPlay, tlv.Type())
	require.Equal(t, 4, tlv.Len())

	ty, l := tlv.TypeLen()
	assert.Equal(t, int(CMDPlay), ty)
	assert.Equal(t, 4, l)
	assert.Equal(t, 2, len(tlv.Value()))
	assert.Equal(t, string([]byte{CMDPlay, 4, 0, 0}), tlv.Str())
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
