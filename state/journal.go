// Copyright 2016 The go-ethereum Authors
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

package state

import (
	"math/big"

	"github.com/dogechain-lab/dogechain/crypto"
	"github.com/dogechain-lab/dogechain/types"
)

// journalEntry is a modification entry in the state change journal that can be
// reverted on demand.
type journalEntry interface {
	// revert undoes the changes introduced by this journal entry.
	revert(*Txn)

	// dirtied returns the Ethereum address modified by this journal entry.
	dirtied() *types.Address
}

// journal contains the list of state modifications applied since the last state
// commit. These are tracked to be able to be reverted in the case of an execution
// exception or request for reversal.
type journal struct {
	entries []journalEntry        // Current changes tracked by the journal
	dirties map[types.Address]int // Dirty accounts and the number of changes
}

// newJournal creates a new initialized journal.
func newJournal() *journal {
	return &journal{
		dirties: make(map[types.Address]int),
	}
}

// append inserts a new modification entry to the end of the change journal.
func (j *journal) append(entry journalEntry) {
	j.entries = append(j.entries, entry)

	if addr := entry.dirtied(); addr != nil {
		j.dirties[*addr]++
	}
}

// revert undoes a batch of journalled modifications along with any reverted
// dirty handling too.
func (j *journal) revert(txn *Txn, snapshot int) {
	for i := len(j.entries) - 1; i >= snapshot; i-- {
		// Undo the changes made by the operation
		j.entries[i].revert(txn)

		// Drop any dirty tracking induced by the change
		if addr := j.entries[i].dirtied(); addr != nil {
			if j.dirties[*addr]--; j.dirties[*addr] == 0 {
				delete(j.dirties, *addr)
			}
		}
	}

	j.entries = j.entries[:snapshot]
}

// length returns the current number of entries in the journal.
func (j *journal) length() int {
	return len(j.entries)
}

type (
	// Changes to the account trie.
	resetObjectChange struct {
		prev         *stateObject
		prevdestruct bool
	}

	suicideChange struct {
		account     *types.Address
		prev        bool // whether account had already suicided
		prevbalance *big.Int
	}

	// Changes to individual accounts.
	balanceChange struct {
		account *types.Address
		prev    *big.Int
	}

	nonceChange struct {
		account *types.Address
		prev    uint64
	}

	codeChange struct {
		account            *types.Address
		prevcode, prevhash []byte
	}
)

func (ch resetObjectChange) revert(t *Txn) {
	if !ch.prevdestruct && t.snap != nil {
		delete(t.snapDestructs, ch.prev.AddressHash())
	}
}

func (ch resetObjectChange) dirtied() *types.Address {
	return nil
}

func addressHash(addr *types.Address) types.Hash {
	return crypto.Keccak256Hash(addr.Bytes())
}

func (ch suicideChange) revert(t *Txn) {
	obj, _ := t.getStateObject(*ch.account)
	if obj == nil {
		delete(t.snapAccounts, addressHash(ch.account))
	} else {
		obj.suicide = ch.prev
		// balance
		obj.setBalance(ch.prevbalance)
		// revert journaled account
		t.updateSnapAccount(obj)
	}
}

func (ch suicideChange) dirtied() *types.Address {
	return ch.account
}

func (ch balanceChange) revert(t *Txn) {
	obj, _ := t.getStateObject(*ch.account)
	if obj == nil {
		delete(t.snapAccounts, addressHash(ch.account))
	} else {
		// balance
		obj.setBalance(ch.prev)
		// revert journaled account
		t.updateSnapAccount(obj)
	}
}

func (ch balanceChange) dirtied() *types.Address {
	return ch.account
}

func (ch nonceChange) revert(t *Txn) {
	obj, _ := t.getStateObject(*ch.account)
	if obj == nil {
		delete(t.snapAccounts, addressHash(ch.account))
	} else {
		// nonce
		obj.setNonce(ch.prev)
		// revert journaled account
		t.updateSnapAccount(obj)
	}
}

func (ch nonceChange) dirtied() *types.Address {
	return ch.account
}

func (ch codeChange) revert(t *Txn) {
	obj, _ := t.getStateObject(*ch.account)
	if obj == nil {
		delete(t.snapAccounts, addressHash(ch.account))
	} else {
		// code
		obj.setCode(types.BytesToHash(ch.prevhash), ch.prevcode)
		// revert journaled account
		t.updateSnapAccount(obj)
	}
}

func (ch codeChange) dirtied() *types.Address {
	return ch.account
}
