package snapshot

import (
	"bytes"

	"github.com/dogechain-lab/dogechain/helper/kvdb"
	"github.com/dogechain-lab/dogechain/types"
)

const journalVersion uint64 = 0

// journalGenerator is a disk layer entry containing the generator progress marker.
type journalGenerator struct {
	// Indicator that whether the database was in progress of being wiped.
	// It's deprecated but keep it here for background compatibility.
	Wiping bool

	Done     bool // Whether the generator finished creating the snapshot
	Marker   []byte
	Accounts uint64
	Slots    uint64
	Storage  uint64
}

// Journal writes the memory layer contents into a buffer to be stored in the
// database as the snapshot journal.
func (dl *diffLayer) Journal(buffer *bytes.Buffer) (types.Hash, error) {
	return types.ZeroHash, nil
}

// Journal terminates any in-progress snapshot generation, also implicitly pushing
// the progress into the database.
func (dl *diskLayer) Journal(buffer *bytes.Buffer) (types.Hash, error) {
	return types.ZeroHash, nil
}

// loadSnapshot loads a pre-existing state snapshot backed by a key-value store.
func loadSnapshot(
	diskdb kvdb.KVBatchStorage,
	root types.Hash,
	cache int,
	recovery bool,
	noBuild bool,
) (snapshot, bool, error) {
	return nil, false, nil
}
