package snapshot

import (
	"bytes"

	"github.com/dogechain-lab/dogechain/helper/kvdb"
	"github.com/dogechain-lab/dogechain/types"
)

const journalVersion uint64 = 0

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
