package stypes

import (
	"fmt"
	"math/big"

	"github.com/dogechain-lab/dogechain/types"
	"github.com/dogechain-lab/fastrlp"
)

// Account is the account reference in the ethereum state
type Account struct {
	Nonce       uint64
	Balance     *big.Int
	StorageRoot types.Hash // storage root
	CodeHash    []byte
}

func (a *Account) MarshalWith(ar *fastrlp.Arena) *fastrlp.Value {
	v := ar.NewArray()
	v.Set(ar.NewUint(a.Nonce))
	v.Set(ar.NewBigInt(a.Balance))
	v.Set(ar.NewBytes(a.StorageRoot.Bytes()))
	v.Set(ar.NewBytes(a.CodeHash))

	return v
}

var accountParserPool fastrlp.ParserPool

func (a *Account) UnmarshalRlp(b []byte) error {
	p := accountParserPool.Get()
	defer accountParserPool.Put(p)

	v, err := p.Parse(b)
	if err != nil {
		return err
	}

	elems, err := v.GetElems()

	if err != nil {
		return err
	}

	if len(elems) < 4 {
		return fmt.Errorf("incorrect number of elements to decode account, expected at least 4 but found %d",
			len(elems))
	}

	// nonce
	if a.Nonce, err = elems[0].GetUint64(); err != nil {
		return err
	}
	// balance
	if a.Balance == nil {
		a.Balance = new(big.Int)
	}

	if err = elems[1].GetBigInt(a.Balance); err != nil {
		return err
	}
	// root
	if err = elems[2].GetHash(a.StorageRoot[:]); err != nil {
		return err
	}
	// codeHash
	if a.CodeHash, err = elems[3].GetBytes(a.CodeHash[:0]); err != nil {
		return err
	}

	return nil
}

func (a *Account) String() string {
	return fmt.Sprintf("%d %s", a.Nonce, a.Balance.String())
}

func (a *Account) Copy() *Account {
	aa := new(Account)

	if a.Balance == nil {
		aa.Balance = new(big.Int)
	} else {
		aa.Balance = big.NewInt(0).SetBytes(a.Balance.Bytes())
	}

	aa.Nonce = a.Nonce
	aa.CodeHash = a.CodeHash
	aa.StorageRoot = a.StorageRoot

	return aa
}
