package schema

import "github.com/dogechain-lab/dogechain/types"

var (
	// codePrefix is the code prefix for leveldb
	CodePrefix = []byte("code")
)

func CodeKey(codeHash types.Hash) []byte {
	return append(CodePrefix, codeHash.Bytes()...)
}
