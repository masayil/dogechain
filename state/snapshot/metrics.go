package snapshot

import (
	"strings"

	"github.com/dogechain-lab/dogechain/helper/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	_subsystemID = "snapshot"
)

type generateMetricContext struct {
	generateSeconds metrics.DurationContext
	accountProve    metrics.DurationContext
	accountTrieRead metrics.DurationContext
	accountSnapRead metrics.DurationContext
	accountWrite    metrics.DurationContext
	storageProve    metrics.DurationContext
	storageTrieRead metrics.DurationContext
	storageSnapRead metrics.DurationContext
	storageWrite    metrics.DurationContext
	storageClean    metrics.DurationContext
}

func (ctx *generateMetricContext) Start() {
	ctx.generateSeconds.Start()
}

type generateMetrics struct {
	generatedAccountCount prometheus.Counter
	recoveredAccountCount prometheus.Counter
	wipedAccountCount     prometheus.Counter
	missallAccountCount   prometheus.Counter
	generatedStorageCount prometheus.Counter
	recoveredStorageCount prometheus.Counter
	wipedStorageCount     prometheus.Counter
	missallStorageCount   prometheus.Counter
	// danglingStorageSize       metrics.TotalCountHistogram
	successfulRangeProofCount prometheus.Counter
	failedRangeProofCount     prometheus.Counter
	generateSeconds           prometheus.Gauge

	// accountProveNanoseconds measures time spent on the account proving
	accountProveNanoseconds prometheus.Histogram
	// accountTrieReadNanoseconds measures time spent on the account trie iteration
	accountTrieReadNanoseconds prometheus.Histogram
	// accountSnapReadNanoseconds measures time spent on the snapshot account iteration
	accountSnapReadNanoseconds prometheus.Histogram
	// accountWriteNanoseconds measures time spent on writing/updating/deleting accounts
	accountWriteNanoseconds prometheus.Histogram
	// storageProveNanoseconds measures time spent on storage proving
	storageProveNanoseconds prometheus.Histogram
	// storageTrieReadNanoseconds measures time spent on the storage trie iteration
	storageTrieReadNanoseconds prometheus.Histogram
	// storageSnapReadNanoseconds measures time spent on the snapshot storage iteration
	storageSnapReadNanoseconds prometheus.Histogram
	// storageWriteNanoseconds measures time spent on writing/updating storages
	storageWriteNanoseconds prometheus.Histogram
	// storageCleanNanoseconds measures time spent on deleting storages
	storageCleanNanoseconds prometheus.Histogram
}

func newGenerateMetrics(namespace string, constLabels prometheus.Labels) *generateMetrics {
	var (
		generatedAccountCount = newCounter(namespace, "generate_generated_account_count", constLabels)
		recoveredAccountCount = newCounter(namespace, "generate_recovered_account_count", constLabels)
		wipedAccountCount     = newCounter(namespace, "generate_wiped_account_count", constLabels)
		missallAccountCount   = newCounter(namespace, "generate_missall_account_count", constLabels)
		generatedStorageCount = newCounter(namespace, "generate_generated_storage_count", constLabels)
		recoveredStorageCount = newCounter(namespace, "generate_recovered_storage_count", constLabels)
		wipedStorageCount     = newCounter(namespace, "generate_wiped_storage_count", constLabels)
		missallStorageCount   = newCounter(namespace, "generate_missall_storage_count", constLabels)
		// danglingStorageSize        = newHistogram(namespace, "generate_dangling_storage_size", constLabels)
		successfulRangeProofCount = newCounter(namespace, "generate_successful_range_proof_count", constLabels)
		failedRangeProofCount     = newCounter(namespace, "generate_failed_range_proof_count", constLabels)
		generateSeconds           = newGauge(namespace, "generate_generate_seconds", constLabels)
		// all nanoseconds metrics
		accountProveNanoseconds    = newHistogram(namespace, "generate_account_prove_nanoseconds", constLabels)
		accountTrieReadNanoSeconds = newHistogram(namespace, "generate_account_trie_read_nanoseconds", constLabels)
		accountSnapReadNanoseconds = newHistogram(namespace, "generate_account_snap_read_nanoseconds", constLabels)
		accountWriteNanoseconds    = newHistogram(namespace, "generate_account_write_nanoseconds", constLabels)
		storageProveNanoseconds    = newHistogram(namespace, "generate_storage_prove_nanoseconds", constLabels)
		storageTrieReadNanoseconds = newHistogram(namespace, "generate_storage_trie_read_nanoseconds", constLabels)
		storageSnapReadNanoseconds = newHistogram(namespace, "generate_storage_snap_read_nanoseconds", constLabels)
		storageWriteNanoseconds    = newHistogram(namespace, "generate_storage_write_nanoseconds", constLabels)
		storageCleanNanoseconds    = newHistogram(namespace, "generate_storage_clean_nanoseconds", constLabels)
	)

	prometheus.MustRegister(generatedAccountCount)
	prometheus.MustRegister(recoveredAccountCount)
	prometheus.MustRegister(wipedAccountCount)
	prometheus.MustRegister(missallAccountCount)
	prometheus.MustRegister(generatedStorageCount)
	prometheus.MustRegister(recoveredStorageCount)
	prometheus.MustRegister(wipedStorageCount)
	prometheus.MustRegister(missallStorageCount)
	// prometheus.MustRegister(danglingStorageSize)
	prometheus.MustRegister(successfulRangeProofCount)
	prometheus.MustRegister(failedRangeProofCount)
	prometheus.MustRegister(generateSeconds)
	prometheus.MustRegister(accountProveNanoseconds)
	prometheus.MustRegister(accountTrieReadNanoSeconds)
	prometheus.MustRegister(accountSnapReadNanoseconds)
	prometheus.MustRegister(accountWriteNanoseconds)
	prometheus.MustRegister(storageProveNanoseconds)
	prometheus.MustRegister(storageTrieReadNanoseconds)
	prometheus.MustRegister(storageSnapReadNanoseconds)
	prometheus.MustRegister(storageWriteNanoseconds)
	prometheus.MustRegister(storageCleanNanoseconds)

	return &generateMetrics{
		generatedAccountCount: generatedAccountCount,
		recoveredAccountCount: recoveredAccountCount,
		wipedAccountCount:     wipedAccountCount,
		missallAccountCount:   missallAccountCount,
		generatedStorageCount: generatedStorageCount,
		recoveredStorageCount: recoveredStorageCount,
		wipedStorageCount:     wipedStorageCount,
		missallStorageCount:   missallStorageCount,
		// danglingStorageSize:        metrics.NewTotalCounterHistogram(danglingStorageSize),
		successfulRangeProofCount:  successfulRangeProofCount,
		failedRangeProofCount:      failedRangeProofCount,
		generateSeconds:            generateSeconds,
		accountProveNanoseconds:    accountProveNanoseconds,
		accountTrieReadNanoseconds: accountTrieReadNanoSeconds,
		accountSnapReadNanoseconds: accountSnapReadNanoseconds,
		accountWriteNanoseconds:    accountWriteNanoseconds,
		storageProveNanoseconds:    storageProveNanoseconds,
		storageTrieReadNanoseconds: storageTrieReadNanoseconds,
		storageSnapReadNanoseconds: storageSnapReadNanoseconds,
		storageWriteNanoseconds:    storageWriteNanoseconds,
		storageCleanNanoseconds:    storageCleanNanoseconds,
	}
}

func nilGenerateMetrics() *generateMetrics {
	return &generateMetrics{}
}

func (m *generateMetrics) Context() *generateMetricContext {
	return &generateMetricContext{
		generateSeconds: metrics.NewDurationContextWithUnit(metrics.DurationSecond),
		accountProve:    metrics.NewDurationContextWithUnit(metrics.DurationNanosecond),
		accountTrieRead: metrics.NewDurationContextWithUnit(metrics.DurationNanosecond),
		accountSnapRead: metrics.NewDurationContextWithUnit(metrics.DurationNanosecond),
		accountWrite:    metrics.NewDurationContextWithUnit(metrics.DurationNanosecond),
		storageProve:    metrics.NewDurationContextWithUnit(metrics.DurationNanosecond),
		storageTrieRead: metrics.NewDurationContextWithUnit(metrics.DurationNanosecond),
		storageSnapRead: metrics.NewDurationContextWithUnit(metrics.DurationNanosecond),
		storageWrite:    metrics.NewDurationContextWithUnit(metrics.DurationNanosecond),
		storageClean:    metrics.NewDurationContextWithUnit(metrics.DurationNanosecond),
	}
}

type Metrics struct {
	*generateMetrics

	cleanAccountHitCount  prometheus.Counter
	cleanAccountMissCount prometheus.Counter
	cleanAccountInexCount prometheus.Counter
	cleanAccountReadSize  prometheus.Histogram
	cleanAccountWriteSize prometheus.Histogram

	cleanStorageHitCount  prometheus.Counter
	cleanStorageMissCount prometheus.Counter
	cleanStorageInexCount prometheus.Counter
	cleanStorageReadSize  prometheus.Histogram
	cleanStorageWriteSize prometheus.Histogram

	dirtyAccountHitCount  prometheus.Counter
	dirtyAccountMissCount prometheus.Counter
	dirtyAccountInexCount prometheus.Counter
	dirtyAccountReadSize  prometheus.Histogram
	dirtyAccountWriteSize prometheus.Histogram

	dirtyStorageHitCount  prometheus.Counter
	dirtyStorageMissCount prometheus.Counter
	dirtyStorageInexCount prometheus.Counter
	dirtyStorageReadSize  prometheus.Histogram
	dirtyStorageWriteSize prometheus.Histogram

	dirtyAccountHitDepth prometheus.Histogram
	dirtyStorageHitDepth prometheus.Histogram

	flushAccountItemCount prometheus.Counter
	flushAccountSize      prometheus.Histogram
	flushStorageItemCount prometheus.Counter
	flushStorageSize      prometheus.Histogram

	bloomIndexNanoseconds prometheus.Histogram
	bloomErrorCount       prometheus.Gauge

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
		cleanAccountHitCount:      newCounter(namespace, "clean_account_hit_count", constLabels),
		cleanAccountMissCount:     newCounter(namespace, "clean_account_miss_count", constLabels),
		cleanAccountInexCount:     newCounter(namespace, "clean_account_inex_count", constLabels),
		cleanAccountReadSize:      newHistogram(namespace, "clean_account_read_size", constLabels),
		cleanAccountWriteSize:     newHistogram(namespace, "clean_account_write_size", constLabels),
		cleanStorageHitCount:      newCounter(namespace, "clean_storage_hit_count", constLabels),
		cleanStorageMissCount:     newCounter(namespace, "clean_storage_miss_count", constLabels),
		cleanStorageInexCount:     newCounter(namespace, "clean_storage_inex_count", constLabels),
		cleanStorageReadSize:      newHistogram(namespace, "clean_storage_read_size", constLabels),
		cleanStorageWriteSize:     newHistogram(namespace, "clean_storage_write_size", constLabels),
		dirtyAccountHitCount:      newCounter(namespace, "dirty_account_hit_count", constLabels),
		dirtyAccountMissCount:     newCounter(namespace, "dirty_account_miss_count", constLabels),
		dirtyAccountInexCount:     newCounter(namespace, "dirty_account_inex_count", constLabels),
		dirtyAccountReadSize:      newHistogram(namespace, "dirty_account_read_size", constLabels),
		dirtyAccountWriteSize:     newHistogram(namespace, "dirty_account_write_size", constLabels),
		dirtyStorageHitCount:      newCounter(namespace, "dirty_storage_hit_count", constLabels),
		dirtyStorageMissCount:     newCounter(namespace, "dirty_storage_miss_count", constLabels),
		dirtyStorageInexCount:     newCounter(namespace, "dirty_storage_inex_count", constLabels),
		dirtyStorageReadSize:      newHistogram(namespace, "dirty_storage_read_size", constLabels),
		dirtyStorageWriteSize:     newHistogram(namespace, "dirty_storage_write_size", constLabels),
		dirtyAccountHitDepth:      newHistogram(namespace, "dirty_account_hit_depth", constLabels),
		dirtyStorageHitDepth:      newHistogram(namespace, "dirty_storage_hit_depth", constLabels),
		flushAccountItemCount:     newCounter(namespace, "flush_account_item_count", constLabels),
		flushAccountSize:          newHistogram(namespace, "flush_account_size", constLabels),
		flushStorageItemCount:     newCounter(namespace, "flush_storage_item_count", constLabels),
		flushStorageSize:          newHistogram(namespace, "flush_storage_size", constLabels),
		bloomIndexNanoseconds:     newHistogram(namespace, "bloom_index_nanoseconds", constLabels),
		bloomErrorCount:           newGauge(namespace, "bloom_error_count", constLabels),
		bloomAccountTrueHitCount:  newCounter(namespace, "bloom_account_true_hit_count", constLabels),
		bloomAccountFalseHitCount: newCounter(namespace, "bloom_account_false_hit_count", constLabels),
		bloomAccountMissCount:     newCounter(namespace, "bloom_account_miss_count", constLabels),
		bloomStorageTrueHitCount:  newCounter(namespace, "bloom_storage_true_hit_count", constLabels),
		bloomStorageFalseHitCount: newCounter(namespace, "bloom_storage_false_hit_count", constLabels),
		bloomStorageMissCount:     newCounter(namespace, "bloom_storage_miss_count", constLabels),
	}

	m.generateMetrics = newGenerateMetrics(namespace, constLabels)

	prometheus.MustRegister(
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
		m.bloomIndexNanoseconds,
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
		Subsystem:   _subsystemID,
		Name:        name,
		Help:        metricName2Help(name),
		ConstLabels: constLabels,
	})
}

func newCounter(namespace, name string, constLabels prometheus.Labels) prometheus.Counter {
	return prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   namespace,
		Subsystem:   _subsystemID,
		Name:        name,
		Help:        metricName2Help(name),
		ConstLabels: constLabels,
	})
}

func newHistogram(namespace, name string, constLabels prometheus.Labels) prometheus.Histogram {
	return prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace:   namespace,
		Subsystem:   _subsystemID,
		Name:        name,
		Help:        metricName2Help(name),
		ConstLabels: constLabels,
	})
}

// NilMetrics will return the non operational snapshot metrics
func NilMetrics() *Metrics {
	return &Metrics{
		generateMetrics: nilGenerateMetrics(),
	}
}
