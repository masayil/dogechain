package server

import (
	"github.com/dogechain-lab/dogechain/helper/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type jsonrpcStoreMetrics struct {
	counter *prometheus.CounterVec
}

// GetNonce api calls
func (m *jsonrpcStoreMetrics) GetNonceInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "GetNonce"}).Inc()
	}
}

// AddTx api calls
func (m *jsonrpcStoreMetrics) AddTxInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "AddTx"}).Inc()
	}
}

// GetPendingTx api calls
func (m *jsonrpcStoreMetrics) GetPendingTxInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "GetPendingTx"}).Inc()
	}
}

// GetAccount api calls
func (m *jsonrpcStoreMetrics) GetAccountInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "GetAccount"}).Inc()
	}
}

// GetGetStorage api calls
func (m *jsonrpcStoreMetrics) GetStorageInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "GetStorage"}).Inc()
	}
}

// GetForksInTime api calls
func (m *jsonrpcStoreMetrics) GetForksInTimeInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "GetForksInTime"}).Inc()
	}
}

// GetCode api calls
func (m *jsonrpcStoreMetrics) GetCodeInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "GetCode"}).Inc()
	}
}

// Header api calls
func (m *jsonrpcStoreMetrics) HeaderInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "Header"}).Inc()
	}
}

// GetHeaderByNumber api calls
func (m *jsonrpcStoreMetrics) GetHeaderByNumberInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "GetHeaderByNumber"}).Inc()
	}
}

// GetHeaderByHash api calls
func (m *jsonrpcStoreMetrics) GetHeaderByHashInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "GetHeaderByHash"}).Inc()
	}
}

// GetBlockByHash api calls
func (m *jsonrpcStoreMetrics) GetBlockByHashInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "GetBlockByHash"}).Inc()
	}
}

// GetBlockByNumber api calls
func (m *jsonrpcStoreMetrics) GetBlockByNumberInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "GetBlockByNumber"}).Inc()
	}
}

// ReadTxLookup api calls
func (m *jsonrpcStoreMetrics) ReadTxLookupInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "ReadTxLookup"}).Inc()
	}
}

// GetReceiptsByHash api calls
func (m *jsonrpcStoreMetrics) GetReceiptsByHashInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "GetReceiptsByHash"}).Inc()
	}
}

// GetAvgGasPrice api calls
func (m *jsonrpcStoreMetrics) GetAvgGasPriceInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "GetAvgGasPrice"}).Inc()
	}
}

// ApplyTxn api calls
func (m *jsonrpcStoreMetrics) ApplyTxnInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "ApplyTxn"}).Inc()
	}
}

// GetSyncProgression api calls
func (m *jsonrpcStoreMetrics) GetSyncProgressionInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "GetSyncProgression"}).Inc()
	}
}

// StateAtTransaction api calls
func (m *jsonrpcStoreMetrics) StateAtTransactionInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "StateAtTransaction"}).Inc()
	}
}

// PeerCount api calls
func (m *jsonrpcStoreMetrics) PeerCountInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "PeerCount"}).Inc()
	}
}

// GetTxs api calls
func (m *jsonrpcStoreMetrics) GetTxsInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "GetTxs"}).Inc()
	}
}

// GetCapacity api calls
func (m *jsonrpcStoreMetrics) GetCapacityInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "GetCapacity"}).Inc()
	}
}

// SubscribeEvents api calls
func (m *jsonrpcStoreMetrics) SubscribeEventsInc() {
	if m.counter != nil {
		m.counter.With(prometheus.Labels{"method": "SubscribeEvents"}).Inc()
	}
}

// NewJSONRPCStoreMetrics return the JSONRPCStore metrics instance
func NewJSONRPCStoreMetrics(namespace string, labelsWithValues ...string) *jsonrpcStoreMetrics {
	constLabels := metrics.ParseLables(labelsWithValues...)

	m := &jsonrpcStoreMetrics{
		counter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "jsonrpc_store",
			Name:        "api_call_counter",
			Help:        "api call counter",
			ConstLabels: constLabels,
		}, []string{"method"}),
	}

	prometheus.MustRegister(m.counter)

	return m
}

// JSONRPCStoreNilMetrics will return the non operational jsonrpc metrics
func JSONRPCStoreNilMetrics() *jsonrpcStoreMetrics {
	return &jsonrpcStoreMetrics{
		counter: nil,
	}
}
