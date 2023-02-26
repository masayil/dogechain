package utils

import (
	"github.com/dogechain-lab/dogechain/types"
)

func StorageBytesToHash(v []byte) (types.Hash, error) {
	if len(v) == 0 {
		return types.Hash{}, nil
	}

	vv, err := types.RlpUnmarshal(v)
	if err != nil {
		return types.Hash{}, err
	}

	if vv == nil {
		return types.Hash{}, nil
	}

	res := []byte{}
	if res, err = vv.GetBytes(res[:0]); err != nil {
		return types.Hash{}, err
	}

	return types.BytesToHash(res), nil
}
