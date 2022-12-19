package snapshot

import (
	"bytes"

	"github.com/dogechain-lab/dogechain/types"
)

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
