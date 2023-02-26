package utils

import (
	"testing"

	"github.com/dogechain-lab/dogechain/types"
	"github.com/stretchr/testify/assert"
)

func TestStorageBytesToHash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		val  []byte
		want types.Hash
	}{
		{nil, types.Hash{}},
		{[]byte{0x24}, types.StringToHash("0x24")},
		{[]byte{0x82, 0x01, 0x02}, types.StringToHash("0x0102")},
	}

	for _, tt := range tests {
		actual, err := StorageBytesToHash(tt.val)
		assert.NoError(t, err)
		assert.Equal(t, tt.want, actual)
	}
}
