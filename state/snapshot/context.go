// Copyright 2022 The go-ethereum Authors
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
	"encoding/binary"
	"math"
	"time"

	"github.com/dogechain-lab/dogechain/helper/kvdb"
	"github.com/dogechain-lab/dogechain/types"
)

const (
	snapAccount = "account" // Identifier of account snapshot generation
	snapStorage = "storage" // Identifier of storage snapshot generation
)

// generatorStats is a collection of statistics gathered by the snapshot generator
// for logging purposes.
type generatorStats struct {
	origin   uint64            // Origin prefix where generation started
	start    time.Time         // Timestamp when generation started
	accounts uint64            // Number of accounts indexed(generated or recovered)
	slots    uint64            // Number of storage slots indexed(generated or recovered)
	dangling uint64            // Number of dangling storage slots
	storage  types.StorageSize // Total account and storage slot size(generation or recovery)
	logger   kvdb.Logger       // logger
}

// Log creates an contextual log with the given message and the context pulled
// from the internally maintained statistics.
func (gs *generatorStats) Log(msg string, root types.Hash, marker []byte) {
	var ctx []interface{}
	if root != types.ZeroHash {
		ctx = append(ctx, []interface{}{"root", root}...)
	}

	// Figure out whether we're after or within an account
	switch len(marker) {
	case types.HashLength:
		ctx = append(ctx, []interface{}{"at", types.BytesToHash(marker)}...)
	case 2 * types.HashLength:
		ctx = append(ctx, []interface{}{
			"in", types.BytesToHash(marker[:types.HashLength]),
			"at", types.BytesToHash(marker[types.HashLength:]),
		}...)
	}

	// Add the usual measurements
	ctx = append(ctx, []interface{}{
		"accounts", gs.accounts,
		"slots", gs.slots,
		"storage", gs.storage,
		"dangling", gs.dangling,
		"elapsed", types.PrettyDuration(time.Since(gs.start)),
	}...)

	// Calculate the estimated indexing time based on current stats
	if len(marker) > 0 {
		if done := binary.BigEndian.Uint64(marker[:8]) - gs.origin; done > 0 {
			left := math.MaxUint64 - binary.BigEndian.Uint64(marker[:8])

			speed := done/uint64(time.Since(gs.start)/time.Millisecond+1) + 1 // +1s to avoid division by zero
			ctx = append(ctx, []interface{}{
				"eta", types.PrettyDuration(time.Duration(left/speed) * time.Millisecond),
			}...)
		}
	}

	gs.logger.Info(msg, ctx...)
}
