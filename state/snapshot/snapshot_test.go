// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package snapshot

import (
	"math/big"
	"math/rand"

	"github.com/dogechain-lab/dogechain/helper/rlp"
	"github.com/dogechain-lab/dogechain/state/stypes"
	"github.com/dogechain-lab/dogechain/types"
)

// randomHash generates a random blob of data and returns it as a hash.
func randomHash() types.Hash {
	var hash types.Hash
	if n, err := rand.Read(hash[:]); n != types.HashLength || err != nil {
		panic(err)
	}

	return hash
}

// randomAccount generates a random account and returns it RLP encoded.
func randomAccount() []byte {
	root := randomHash()

	a := stypes.Account{
		Balance:     big.NewInt(rand.Int63()),
		Nonce:       rand.Uint64(),
		StorageRoot: root,
		CodeHash:    types.EmptyRootHash.Bytes(),
	}

	data, _ := rlp.EncodeToBytes(a)

	return data
}

// randomAccountSet generates a set of random accounts with the given strings as
// the account address hashes.
func randomAccountSet(hashes ...string) map[types.Hash][]byte {
	accounts := make(map[types.Hash][]byte)
	for _, hash := range hashes {
		accounts[types.StringToHash(hash)] = randomAccount()
	}

	return accounts
}

// randomStorageSet generates a set of random slots with the given strings as
// the slot addresses.
func randomStorageSet(accounts []string, hashes [][]string, nilStorage [][]string) map[types.Hash]map[types.Hash][]byte {
	storages := make(map[types.Hash]map[types.Hash][]byte)

	for index, account := range accounts {
		storages[types.StringToHash(account)] = make(map[types.Hash][]byte)

		if index < len(hashes) {
			hashes := hashes[index]
			for _, hash := range hashes {
				storages[types.StringToHash(account)][types.StringToHash(hash)] = randomHash().Bytes()
			}
		}

		if index < len(nilStorage) {
			nils := nilStorage[index]
			for _, hash := range nils {
				storages[types.StringToHash(account)][types.StringToHash(hash)] = nil
			}
		}
	}

	return storages
}
