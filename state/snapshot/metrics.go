package snapshot

import (
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
}

// GetPrometheusMetrics return the snapshot metrics instance
func GetPrometheusMetrics(namespace string, trackingIOTimer bool, labelsWithValues ...string) *Metrics {
	constLabels := metrics.ParseLables(labelsWithValues...)

	m := &Metrics{
		trackingIOTimer: trackingIOTimer,
		generatedAccountCount: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "generated_account_count",
			Help:        "generated account count",
			ConstLabels: constLabels,
		}),
		recoveredAccountCount: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "recovered_account_count",
			Help:        "recovered account count",
			ConstLabels: constLabels,
		}),
		wipedAccountCount: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "wiped_account_count",
			Help:        "wiped account count",
			ConstLabels: constLabels,
		}),
		missallAccountCount: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "missall_account_count",
			Help:        "missall account count",
			ConstLabels: constLabels,
		}),
		generatedStorageSize: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "generated_storage_size",
			Help:        "generated storage size",
			ConstLabels: constLabels,
		}),
		recoveredStorageSize: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "recovered_storage_size",
			Help:        "recovered storage size",
			ConstLabels: constLabels,
		}),
		wipedStorageSize: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "wiped_storage_size",
			Help:        "wiped storage size",
			ConstLabels: constLabels,
		}),
		missallStorageCount: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "missall_storage_count",
			Help:        "missall storage count",
			ConstLabels: constLabels,
		}),
		danglingStorageCount: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "dangling_storage_count",
			Help:        "dangling storage count",
			ConstLabels: constLabels,
		}),
		successfulRangeProofCount: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "successful_range_proof_count",
			Help:        "successful range proof count",
			ConstLabels: constLabels,
		}),
		failedRangeProofCount: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "failed_range_proof_count",
			Help:        "failed range proof count",
			ConstLabels: constLabels,
		}),
		accountProveSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "account_prove_seconds",
			Help:        "account prove seconds",
			ConstLabels: constLabels,
		}),
		accountTrieReadSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "account_trie_read_seconds",
			Help:        "account trie read seconds",
			ConstLabels: constLabels,
		}),
		accountSnapReadSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "account_snap_read_seconds",
			Help:        "account snap read seconds",
			ConstLabels: constLabels,
		}),
		accountWriteSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "account_write_seconds",
			Help:        "account write seconds",
			ConstLabels: constLabels,
		}),
		storageProveSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "storage_prove_seconds",
			Help:        "storage prove seconds",
			ConstLabels: constLabels,
		}),
		storageTrieReadSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "storage_trie_read_seconds",
			Help:        "storage trie read seconds",
			ConstLabels: constLabels,
		}),
		storageSnapReadSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "storage_snap_read_seconds",
			Help:        "storage snap read seconds",
			ConstLabels: constLabels,
		}),
		storageWriteSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "storage_write_seconds",
			Help:        "storage write seconds",
			ConstLabels: constLabels,
		}),
		storageCleanSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystemID,
			Name:        "storage_clean_seconds",
			Help:        "storage clean seconds",
			ConstLabels: constLabels,
		}),
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
	)

	return m
}

// NilMetrics will return the non operational snapshot metrics
func NilMetrics() *Metrics {
	return &Metrics{
		trackingIOTimer: false,
	}
}
