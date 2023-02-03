package snapshot

import (
	"bytes"
	"math/big"

	"github.com/dogechain-lab/dogechain/helper/rlp"
	"github.com/dogechain-lab/dogechain/types"
)

// Account is a modified version of a state.Account, where the root is replaced
// with a byte slice. This format can be used to represent full-consensus format
// or slim-snapshot format which replaces the empty root and code hash as nil
// byte slice.
type Account struct {
	Nonce    uint64
	Balance  *big.Int
	Root     []byte
	CodeHash []byte
}

// SlimAccount converts a state.Account content into a slim snapshot account
func SlimAccount(nonce uint64, balance *big.Int, root types.Hash, codehash []byte) Account {
	slim := Account{
		Nonce:   nonce,
		Balance: balance,
	}

	if root != types.EmptyRootHash {
		slim.Root = root[:]
	}

	if !bytes.Equal(codehash, types.EmptyCodeHash.Bytes()) {
		slim.CodeHash = codehash
	}

	return slim
}

// SlimAccountRLP converts a state.Account content into a slim snapshot
// version RLP encoded.
func SlimAccountRLP(nonce uint64, balance *big.Int, root types.Hash, codehash []byte) []byte {
	data, err := rlp.EncodeToBytes(SlimAccount(nonce, balance, root, codehash))
	if err != nil {
		panic(err)
	}

	return data
}

// FullAccount decodes the data on the 'slim RLP' format and return
// the consensus format account.
func FullAccount(data []byte) (Account, error) {
	var account Account
	if err := rlp.DecodeBytes(data, &account); err != nil {
		return Account{}, err
	}

	if len(account.Root) == 0 {
		account.Root = types.EmptyRootHash.Bytes()
	}

	if len(account.CodeHash) == 0 {
		account.CodeHash = types.EmptyCodeHash.Bytes()
	}

	return account, nil
}

// FullAccountRLP converts data on the 'slim RLP' format into the full RLP-format.
func FullAccountRLP(data []byte) ([]byte, error) {
	account, err := FullAccount(data)
	if err != nil {
		return nil, err
	}

	return rlp.EncodeToBytes(account)
}
