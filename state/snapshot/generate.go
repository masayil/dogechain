package snapshot

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/dogechain-lab/dogechain/helper/kvdb"
	"github.com/dogechain-lab/dogechain/helper/rawdb"
	"github.com/dogechain-lab/dogechain/helper/rlp"
	"github.com/dogechain-lab/dogechain/types"
)

// generateSnapshot regenerates a brand new snapshot based on an existing state
// database and head block asynchronously. The snapshot is returned immediately
// and generation is continued in the background until done.
func generateSnapshot(
	diskdb kvdb.KVBatchStorage,
	cache int,
	root types.Hash,
	logger kvdb.Logger,
) *diskLayer {
	// Create a new disk layer with an initialized state marker at zero
	var (
		stats     = &generatorStats{start: time.Now()}
		batch     = diskdb.NewBatch()
		genMarker = []byte{} // Initialized but empty!
	)

	// write batch to db and journal
	rawdb.WriteSnapshotRoot(batch, root)
	journalProgress(batch, genMarker, stats, logger)

	if err := batch.Write(); err != nil {
		logger.Error("Failed to write initialized state marker", "err", err)
		os.Exit(1)
	}

	base := &diskLayer{
		diskdb: diskdb,
		// triedb:     triedb,
		root:       root,
		cache:      fastcache.New(cache * 1024 * 1024),
		genMarker:  genMarker,
		genPending: make(chan struct{}),
		genAbort:   make(chan chan *generatorStats),
		logger:     logger,
	}

	go base.generate(stats)

	logger.Debug("Start snapshot generation", "root", root)

	return base
}

// journalProgress persists the generator stats into the database to resume later.
func journalProgress(
	db kvdb.KVWriter,
	marker []byte,
	stats *generatorStats,
	logger kvdb.Logger,
) {
	// Write out the generator marker. Note it's a standalone disk layer generator
	// which is not mixed with journal. It's ok if the generator is persisted while
	// journal is not.
	entry := journalGenerator{
		Done:   marker == nil,
		Marker: marker,
	}

	if stats != nil {
		entry.Accounts = stats.accounts
		entry.Slots = stats.slots
		entry.Storage = uint64(stats.storage)
	}

	blob, err := rlp.EncodeToBytes(entry)
	if err != nil {
		panic(err) // Cannot happen, here to catch dev errors
	}

	var logstr string

	switch {
	case marker == nil:
		logstr = "done"
	case bytes.Equal(marker, []byte{}):
		logstr = "empty"
	case len(marker) == types.HashLength:
		logstr = fmt.Sprintf("%#x", marker)
	default:
		logstr = fmt.Sprintf("%#x:%#x", marker[:types.HashLength], marker[types.HashLength:])
	}

	logger.Debug("Journalled generator progress", "progress", logstr)

	rawdb.WriteSnapshotGenerator(db, blob)
}

// generate is a background thread that iterates over the state and storage tries,
// constructing the state snapshot. All the arguments are purely for statistics
// gathering and logging, since the method surfs the blocks as they arrive, often
// being restarted.
func (dl *diskLayer) generate(stats *generatorStats) {
	var (
		// accMarker []byte
		abort chan *generatorStats
	)

	// if len(dl.genMarker) > 0 { // []byte{} is the start, use nil for that
	// 	accMarker = dl.genMarker[:types.HashLength]
	// }

	stats.Log("Resuming state snapshot generation", dl.root, dl.genMarker)

	// // Initialize the global generator context. The snapshot iterators are
	// // opened at the interrupted position because the assumption is held
	// // that all the snapshot data are generated correctly before the marker.
	// // Even if the snapshot data is updated during the interruption (before
	// // or at the marker), the assumption is still held.
	// // For the account or storage slot at the interruption, they will be
	// // processed twice by the generator(they are already processed in the
	// // last run) but it's fine.
	// ctx := newGeneratorContext(stats, dl.diskdb, accMarker, dl.genMarker)
	// defer ctx.close()

	// if err := generateAccounts(ctx, dl, accMarker); err != nil {
	// 	// Extract the received interruption signal if exists
	// 	if aerr, ok := err.(*abortErr); ok {
	// 		abort = aerr.abort
	// 	}
	// 	// Aborted by internal error, wait the signal
	// 	if abort == nil {
	// 		abort = <-dl.genAbort
	// 	}
	// 	abort <- stats
	// 	return
	// }
	// // Snapshot fully generated, set the marker to nil.
	// // Note even there is nothing to commit, persist the
	// // generator anyway to mark the snapshot is complete.
	// journalProgress(ctx.batch, nil, stats)
	// if err := ctx.batch.Write(); err != nil {
	// 	dl.logger.Error("Failed to flush batch", "err", err)

	// 	abort = <-dl.genAbort
	// 	abort <- stats
	// 	return
	// }
	// ctx.batch.Reset()

	dl.logger.Info("Generated state snapshot",
		"accounts", stats.accounts,
		"slots", stats.slots,
		"storage", stats.storage,
		"dangling", stats.dangling,
		"elapsed", types.PrettyDuration(time.Since(stats.start)),
	)

	dl.lock.Lock()
	dl.genMarker = nil
	close(dl.genPending)
	dl.lock.Unlock()

	// Someone will be looking for us, wait it out
	abort = <-dl.genAbort
	abort <- nil
}
