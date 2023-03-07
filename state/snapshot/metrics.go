package snapshot

import (
	"strings"

	"github.com/dogechain-lab/dogechain/helper/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const subsystemID = "snapshot"

type Metrics struct {
	generatedAccountCount     prometheus.Counter
	recoveredAccountCount     prometheus.Counter
	wipedAccountCount         prometheus.Counter
	missallAccountCount       prometheus.Counter
	generatedStorageSize      prometheus.Counter
	recoveredStorageSize      prometheus.Counter
	wipedStorageSize          prometheus.Counter
	missallStorageCount       prometheus.Counter
	danglingStorageSize       prometheus.Counter
	successfulRangeProofCount prometheus.Counter
	failedRangeProofCount     prometheus.Counter

	// accountProveSeconds measures time spent on the account proving
	accountProveSeconds prometheus.Counter
	// accountTrieReadSeconds measures time spent on the account trie iteration
	accountTrieReadSeconds prometheus.Counter
	// accountSnapReadSeconds measures time spent on the snapshot account iteration
	accountSnapReadSeconds prometheus.Counter
	// accountWriteSeconds measures time spent on writing/updating/deleting accounts
	accountWriteSeconds prometheus.Counter
	// storageProveSeconds measures time spent on storage proving
	storageProveSeconds prometheus.Counter
	// storageTrieReadSeconds measures time spent on the storage trie iteration
	storageTrieReadSeconds prometheus.Counter
	// storageSnapReadSeconds measures time spent on the snapshot storage iteration
	storageSnapReadSeconds prometheus.Counter
	// storageWriteSeconds measures time spent on writing/updating storages
	storageWriteSeconds prometheus.Counter
	// storageCleanSeconds measures time spent on deleting storages
	storageCleanSeconds prometheus.Counter

	cleanAccountHitCount  prometheus.Counter
	cleanAccountMissCount prometheus.Counter
	cleanAccountInexCount prometheus.Counter
	cleanAccountReadSize  prometheus.Counter
	cleanAccountWriteSize prometheus.Counter

	cleanStorageHitCount  prometheus.Counter
	cleanStorageMissCount prometheus.Counter
	cleanStorageInexCount prometheus.Counter
	cleanStorageReadSize  prometheus.Counter
	cleanStorageWriteSize prometheus.Counter

	dirtyAccountHitCount  prometheus.Counter
	dirtyAccountMissCount prometheus.Counter
	dirtyAccountInexCount prometheus.Counter
	dirtyAccountReadSize  prometheus.Counter
	dirtyAccountWriteSize prometheus.Counter

	dirtyStorageHitCount  prometheus.Counter
	dirtyStorageMissCount prometheus.Counter
	dirtyStorageInexCount prometheus.Counter
	dirtyStorageReadSize  prometheus.Counter
	dirtyStorageWriteSize prometheus.Counter

	dirtyAccountHitDepth prometheus.Histogram
	dirtyStorageHitDepth prometheus.Histogram

	flushAccountItemCount prometheus.Counter
	flushAccountSize      prometheus.Counter
	flushStorageItemCount prometheus.Counter
	flushStorageSize      prometheus.Counter

	bloomIndexTimer prometheus.Histogram
	bloomErrorCount prometheus.Gauge

	bloomAccountTrueHitCount  prometheus.Counter
	bloomAccountFalseHitCount prometheus.Counter
	bloomAccountMissCount     prometheus.Counter

	bloomStorageTrueHitCount  prometheus.Counter
	bloomStorageFalseHitCount prometheus.Counter
	bloomStorageMissCount     prometheus.Counter
}

// GetPrometheusMetrics return the snapshot metrics instance
func GetPrometheusMetrics(namespace string, constLabelsWithValues ...string) *Metrics {
	constLabels := metrics.ParseLables(constLabelsWithValues...)

	m := &Metrics{
		generatedAccountCount:     newCounter(namespace, "generated_account_count", constLabels),
		recoveredAccountCount:     newCounter(namespace, "recovered_account_count", constLabels),
		wipedAccountCount:         newCounter(namespace, "wiped_account_count", constLabels),
		missallAccountCount:       newCounter(namespace, "missall_account_count", constLabels),
		generatedStorageSize:      newCounter(namespace, "generated_storage_size", constLabels),
		recoveredStorageSize:      newCounter(namespace, "recovered_storage_size", constLabels),
		wipedStorageSize:          newCounter(namespace, "wiped_storage_size", constLabels),
		missallStorageCount:       newCounter(namespace, "missall_storage_count", constLabels),
		danglingStorageSize:       newCounter(namespace, "dangling_storage_size", constLabels),
		successfulRangeProofCount: newCounter(namespace, "successful_range_proof_count", constLabels),
		failedRangeProofCount:     newCounter(namespace, "failed_range_proof_count", constLabels),
		accountProveSeconds:       newCounter(namespace, "account_prove_seconds", constLabels),
		accountTrieReadSeconds:    newCounter(namespace, "account_trie_read_seconds", constLabels),
		accountSnapReadSeconds:    newCounter(namespace, "account_snap_read_seconds", constLabels),
		accountWriteSeconds:       newCounter(namespace, "account_write_seconds", constLabels),
		storageProveSeconds:       newCounter(namespace, "storage_prove_seconds", constLabels),
		storageTrieReadSeconds:    newCounter(namespace, "storage_trie_read_seconds", constLabels),
		storageSnapReadSeconds:    newCounter(namespace, "storage_snap_read_seconds", constLabels),
		storageWriteSeconds:       newCounter(namespace, "storage_write_seconds", constLabels),
		storageCleanSeconds:       newCounter(namespace, "storage_clean_seconds", constLabels),
		cleanAccountHitCount:      newCounter(namespace, "clean_account_hit_count", constLabels),
		cleanAccountMissCount:     newCounter(namespace, "clean_account_miss_count", constLabels),
		cleanAccountInexCount:     newCounter(namespace, "clean_account_inex_count", constLabels),
		cleanAccountReadSize:      newCounter(namespace, "clean_account_read_size", constLabels),
		cleanAccountWriteSize:     newCounter(namespace, "clean_account_write_size", constLabels),
		cleanStorageHitCount:      newCounter(namespace, "clean_storage_hit_count", constLabels),
		cleanStorageMissCount:     newCounter(namespace, "clean_storage_miss_count", constLabels),
		cleanStorageInexCount:     newCounter(namespace, "clean_storage_inex_count", constLabels),
		cleanStorageReadSize:      newCounter(namespace, "clean_storage_read_size", constLabels),
		cleanStorageWriteSize:     newCounter(namespace, "clean_storage_write_size", constLabels),
		dirtyAccountHitCount:      newCounter(namespace, "dirty_account_hit_count", constLabels),
		dirtyAccountMissCount:     newCounter(namespace, "dirty_account_miss_count", constLabels),
		dirtyAccountInexCount:     newCounter(namespace, "dirty_account_inex_count", constLabels),
		dirtyAccountReadSize:      newCounter(namespace, "dirty_account_read_size", constLabels),
		dirtyAccountWriteSize:     newCounter(namespace, "dirty_account_write_size", constLabels),
		dirtyStorageHitCount:      newCounter(namespace, "dirty_storage_hit_count", constLabels),
		dirtyStorageMissCount:     newCounter(namespace, "dirty_storage_miss_count", constLabels),
		dirtyStorageInexCount:     newCounter(namespace, "dirty_storage_inex_count", constLabels),
		dirtyStorageReadSize:      newCounter(namespace, "dirty_storage_read_size", constLabels),
		dirtyStorageWriteSize:     newCounter(namespace, "dirty_storage_write_size", constLabels),
		dirtyAccountHitDepth:      newHistogram(namespace, "dirty_account_hit_depth", constLabels),
		dirtyStorageHitDepth:      newHistogram(namespace, "dirty_storage_hit_depth", constLabels),
		flushAccountItemCount:     newCounter(namespace, "flush_account_item_count", constLabels),
		flushAccountSize:          newCounter(namespace, "flush_account_size", constLabels),
		flushStorageItemCount:     newCounter(namespace, "flush_storage_item_count", constLabels),
		flushStorageSize:          newCounter(namespace, "flush_storage_size", constLabels),
		bloomIndexTimer:           newHistogram(namespace, "bloom_index_timer", constLabels),
		bloomErrorCount:           newGauge(namespace, "bloom_error_count", constLabels),
		bloomAccountTrueHitCount:  newCounter(namespace, "bloom_account_true_hit_count", constLabels),
		bloomAccountFalseHitCount: newCounter(namespace, "bloom_account_false_hit_count", constLabels),
		bloomAccountMissCount:     newCounter(namespace, "bloom_account_miss_count", constLabels),
		bloomStorageTrueHitCount:  newCounter(namespace, "bloom_storage_true_hit_count", constLabels),
		bloomStorageFalseHitCount: newCounter(namespace, "bloom_storage_false_hit_count", constLabels),
		bloomStorageMissCount:     newCounter(namespace, "bloom_storage_miss_count", constLabels),
	}

	prometheus.MustRegister(
		m.generatedAccountCount,
		m.recoveredAccountCount,
		m.wipedAccountCount,
		m.missallAccountCount,
		m.generatedStorageSize,
		m.recoveredStorageSize,
		m.wipedStorageSize,
		m.missallStorageCount,
		m.danglingStorageSize,
		m.successfulRangeProofCount,
		m.failedRangeProofCount,
		m.accountProveSeconds,
		m.accountTrieReadSeconds,
		m.accountSnapReadSeconds,
		m.accountWriteSeconds,
		m.storageProveSeconds,
		m.storageTrieReadSeconds,
		m.storageSnapReadSeconds,
		m.storageWriteSeconds,
		m.storageCleanSeconds,
		m.cleanAccountHitCount,
		m.cleanAccountMissCount,
		m.cleanAccountInexCount,
		m.cleanAccountReadSize,
		m.cleanAccountWriteSize,
		m.cleanStorageHitCount,
		m.cleanStorageMissCount,
		m.cleanStorageInexCount,
		m.cleanStorageReadSize,
		m.cleanStorageWriteSize,
		m.dirtyAccountHitCount,
		m.dirtyAccountMissCount,
		m.dirtyAccountInexCount,
		m.dirtyAccountReadSize,
		m.dirtyAccountWriteSize,
		m.dirtyStorageHitCount,
		m.dirtyStorageMissCount,
		m.dirtyStorageInexCount,
		m.dirtyStorageReadSize,
		m.dirtyStorageWriteSize,
		m.dirtyAccountHitDepth,
		m.dirtyStorageHitDepth,
		m.flushAccountItemCount,
		m.flushAccountSize,
		m.flushStorageItemCount,
		m.flushStorageSize,
		m.bloomIndexTimer,
		m.bloomErrorCount,
		m.bloomAccountTrueHitCount,
		m.bloomAccountFalseHitCount,
		m.bloomAccountMissCount,
		m.bloomStorageTrueHitCount,
		m.bloomStorageFalseHitCount,
		m.bloomStorageMissCount,
	)

	return m
}

func metricName2Help(name string) string {
	return strings.ReplaceAll(name, "_", " ")
}

func newGauge(namespace, name string, constLabels prometheus.Labels) prometheus.Gauge {
	return prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   namespace,
		Subsystem:   subsystemID,
		Name:        name,
		Help:        metricName2Help(name),
		ConstLabels: constLabels,
	})
}

func newCounter(namespace, name string, constLabels prometheus.Labels) prometheus.Counter {
	return prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   namespace,
		Subsystem:   subsystemID,
		Name:        name,
		Help:        metricName2Help(name),
		ConstLabels: constLabels,
	})
}

func newHistogram(namespace, name string, constLabels prometheus.Labels) prometheus.Histogram {
	return prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace:   namespace,
		Subsystem:   subsystemID,
		Name:        name,
		Help:        metricName2Help(name),
		ConstLabels: constLabels,
	})
}

// NilMetrics will return the non operational snapshot metrics
func NilMetrics() *Metrics {
	return &Metrics{}
}
