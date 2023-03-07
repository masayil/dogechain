package snapshot

import (
	"strings"

	"github.com/dogechain-lab/dogechain/helper/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const subsystemID = "snapshot"

type Metrics struct {
	generatedAccountCount     prometheus.Histogram
	recoveredAccountCount     prometheus.Histogram
	wipedAccountCount         prometheus.Histogram
	missallAccountCount       prometheus.Histogram
	generatedStorageSize      prometheus.Histogram
	recoveredStorageSize      prometheus.Histogram
	wipedStorageSize          prometheus.Histogram
	missallStorageCount       prometheus.Histogram
	danglingStorageCount      prometheus.Histogram
	successfulRangeProofCount prometheus.Histogram
	failedRangeProofCount     prometheus.Histogram

	// duration measurement might be time consuming, ignore it if possible
	trackingIOTimer bool

	// accountProveSeconds measures time spent on the account proving
	accountProveSeconds prometheus.Histogram
	// accountTrieReadSeconds measures time spent on the account trie iteration
	accountTrieReadSeconds prometheus.Histogram
	// accountSnapReadSeconds measures time spent on the snapshot account iteration
	accountSnapReadSeconds prometheus.Histogram
	// accountWriteSeconds measures time spent on writing/updating/deleting accounts
	accountWriteSeconds prometheus.Histogram
	// storageProveSeconds measures time spent on storage proving
	storageProveSeconds prometheus.Histogram
	// storageTrieReadSeconds measures time spent on the storage trie iteration
	storageTrieReadSeconds prometheus.Histogram
	// storageSnapReadSeconds measures time spent on the snapshot storage iteration
	storageSnapReadSeconds prometheus.Histogram
	// storageWriteSeconds measures time spent on writing/updating storages
	storageWriteSeconds prometheus.Histogram
	// storageCleanSeconds measures time spent on deleting storages
	storageCleanSeconds prometheus.Histogram

	cleanAccountHitCount  prometheus.Histogram
	cleanAccountMissCount prometheus.Histogram
	cleanAccountInexCount prometheus.Histogram
	cleanAccountReadSize  prometheus.Histogram
	cleanAccountWriteSize prometheus.Histogram

	cleanStorageHitCount  prometheus.Histogram
	cleanStorageMissCount prometheus.Histogram
	cleanStorageInexCount prometheus.Histogram
	cleanStorageReadSize  prometheus.Histogram
	cleanStorageWriteSize prometheus.Histogram

	dirtyAccountHitCount  prometheus.Histogram
	dirtyAccountMissCount prometheus.Histogram
	dirtyAccountInexCount prometheus.Histogram
	dirtyAccountReadSize  prometheus.Histogram
	dirtyAccountWriteSize prometheus.Histogram

	dirtyStorageHitCount  prometheus.Histogram
	dirtyStorageMissCount prometheus.Histogram
	dirtyStorageInexCount prometheus.Histogram
	dirtyStorageReadSize  prometheus.Histogram
	dirtyStorageWriteSize prometheus.Histogram

	dirtyAccountHitDepth prometheus.Histogram
	dirtyStorageHitDepth prometheus.Histogram

	flushAccountItemCount prometheus.Histogram
	flushAccountSize      prometheus.Histogram
	flushStorageItemCount prometheus.Histogram
	flushStorageSize      prometheus.Histogram

	bloomIndexTimer prometheus.Gauge
	bloomErrorCount prometheus.Gauge

	bloomAccountTrueHitCount  prometheus.Histogram
	bloomAccountFalseHitCount prometheus.Histogram
	bloomAccountMissCount     prometheus.Histogram

	bloomStorageTrueHitCount  prometheus.Histogram
	bloomStorageFalseHitCount prometheus.Histogram
	bloomStorageMissCount     prometheus.Histogram
}

// GetPrometheusMetrics return the snapshot metrics instance
func GetPrometheusMetrics(namespace string, trackingIOTimer bool, commonLabels ...string) *Metrics {
	constLabels := metrics.ParseLables(commonLabels...)

	m := &Metrics{
		trackingIOTimer:           trackingIOTimer,
		generatedAccountCount:     newHistogram(namespace, "generated_account_count", constLabels),
		recoveredAccountCount:     newHistogram(namespace, "recovered_account_count", constLabels),
		wipedAccountCount:         newHistogram(namespace, "wiped_account_count", constLabels),
		missallAccountCount:       newHistogram(namespace, "missall_account_count", constLabels),
		generatedStorageSize:      newHistogram(namespace, "generated_storage_size", constLabels),
		recoveredStorageSize:      newHistogram(namespace, "recovered_storage_size", constLabels),
		wipedStorageSize:          newHistogram(namespace, "wiped_storage_size", constLabels),
		missallStorageCount:       newHistogram(namespace, "missall_storage_count", constLabels),
		danglingStorageCount:      newHistogram(namespace, "dangling_storage_count", constLabels),
		successfulRangeProofCount: newHistogram(namespace, "successful_range_proof_count", constLabels),
		failedRangeProofCount:     newHistogram(namespace, "failed_range_proof_count", constLabels),
		accountProveSeconds:       newHistogram(namespace, "account_prove_seconds", constLabels),
		accountTrieReadSeconds:    newHistogram(namespace, "account_trie_read_seconds", constLabels),
		accountSnapReadSeconds:    newHistogram(namespace, "account_snap_read_seconds", constLabels),
		accountWriteSeconds:       newHistogram(namespace, "account_write_seconds", constLabels),
		storageProveSeconds:       newHistogram(namespace, "storage_prove_seconds", constLabels),
		storageTrieReadSeconds:    newHistogram(namespace, "storage_trie_read_seconds", constLabels),
		storageSnapReadSeconds:    newHistogram(namespace, "storage_snap_read_seconds", constLabels),
		storageWriteSeconds:       newHistogram(namespace, "storage_write_seconds", constLabels),
		storageCleanSeconds:       newHistogram(namespace, "storage_clean_seconds", constLabels),
		cleanAccountHitCount:      newHistogram(namespace, "clean_account_hit_count", constLabels),
		cleanAccountMissCount:     newHistogram(namespace, "clean_account_miss_count", constLabels),
		cleanAccountInexCount:     newHistogram(namespace, "clean_account_inex_count", constLabels),
		cleanAccountReadSize:      newHistogram(namespace, "clean_account_read_size", constLabels),
		cleanAccountWriteSize:     newHistogram(namespace, "clean_account_write_size", constLabels),
		cleanStorageHitCount:      newHistogram(namespace, "clean_storage_hit_count", constLabels),
		cleanStorageMissCount:     newHistogram(namespace, "clean_storage_miss_count", constLabels),
		cleanStorageInexCount:     newHistogram(namespace, "clean_storage_inex_count", constLabels),
		cleanStorageReadSize:      newHistogram(namespace, "clean_storage_read_size", constLabels),
		cleanStorageWriteSize:     newHistogram(namespace, "clean_storage_write_size", constLabels),
		dirtyAccountHitCount:      newHistogram(namespace, "dirty_account_hit_count", constLabels),
		dirtyAccountMissCount:     newHistogram(namespace, "dirty_account_miss_count", constLabels),
		dirtyAccountInexCount:     newHistogram(namespace, "dirty_account_inex_count", constLabels),
		dirtyAccountReadSize:      newHistogram(namespace, "dirty_account_read_size", constLabels),
		dirtyAccountWriteSize:     newHistogram(namespace, "dirty_account_write_size", constLabels),
		dirtyStorageHitCount:      newHistogram(namespace, "dirty_storage_hit_count", constLabels),
		dirtyStorageMissCount:     newHistogram(namespace, "dirty_storage_miss_count", constLabels),
		dirtyStorageInexCount:     newHistogram(namespace, "dirty_storage_inex_count", constLabels),
		dirtyStorageReadSize:      newHistogram(namespace, "dirty_storage_read_size", constLabels),
		dirtyStorageWriteSize:     newHistogram(namespace, "dirty_storage_write_size", constLabels),
		dirtyAccountHitDepth:      newHistogram(namespace, "dirty_account_hit_depth", constLabels),
		dirtyStorageHitDepth:      newHistogram(namespace, "dirty_storage_hit_depth", constLabels),
		flushAccountItemCount:     newHistogram(namespace, "flush_account_item_count", constLabels),
		flushAccountSize:          newHistogram(namespace, "flush_account_size", constLabels),
		flushStorageItemCount:     newHistogram(namespace, "flush_storage_item_count", constLabels),
		flushStorageSize:          newHistogram(namespace, "flush_storage_size", constLabels),
		bloomIndexTimer:           newGauge(namespace, "bloom_index_timer", constLabels),
		bloomErrorCount:           newGauge(namespace, "bloom_error_count", constLabels),
		bloomAccountTrueHitCount:  newHistogram(namespace, "bloom_account_true_hit_count", constLabels),
		bloomAccountFalseHitCount: newHistogram(namespace, "bloom_account_false_hit_count", constLabels),
		bloomAccountMissCount:     newHistogram(namespace, "bloom_account_miss_count", constLabels),
		bloomStorageTrueHitCount:  newHistogram(namespace, "bloom_storage_true_hit_count", constLabels),
		bloomStorageFalseHitCount: newHistogram(namespace, "bloom_storage_false_hit_count", constLabels),
		bloomStorageMissCount:     newHistogram(namespace, "bloom_storage_miss_count", constLabels),
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
		m.danglingStorageCount,
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
	return &Metrics{
		trackingIOTimer: false,
	}
}
