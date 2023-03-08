package metrics

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type DurationUnit int

const (
	DurationSecond DurationUnit = iota
	DurationMillisecond
	DurationMicrosecond
	DurationNanosecond
)

// DurationContext is use for accumulating duration
type DurationContext interface {
	SetUnit(DurationUnit) // set time counting units
	Start()               // start timer, exclusive with Add
	Add(time.Duration)    // increase duration, exclusive with Start, should be thread safe
	Duration() float64    // sum up duration on base unit
}

// CounterContext is use for accumulating value (count or size)
type CounterContext interface {
	Add(int64)    // add value count(or size), should be thread safe
	Count() int64 // number of times
	Value() int64 // sum up value
}

// CumulativeDurationFn represents duration context handling for histogram
// metric.
type CumulativeDurationFn func(DurationContext)

// DurationHistogram is histogram used for duration.
type DurationHistogram interface {
	TimeAccumulator() CumulativeDurationFn
}

// CountTotalFn is used to count total the context holding.
type CountTotalFn func(CounterContext)

// TotalCountHistogram is histogram used for total count (or size).
type TotalCountHistogram interface {
	CountAccumulator() CountTotalFn
}

// standard duration context
type standardDurationContext struct {
	unit      DurationUnit
	starttime time.Time
	tmp       int64
}

func NewDurationContextWithUnit(u DurationUnit) DurationContext {
	return &standardDurationContext{unit: u}
}
func (ctx *standardDurationContext) SetUnit(u DurationUnit) { ctx.unit = u }
func (ctx *standardDurationContext) Start()                 { ctx.starttime = time.Now() }
func (ctx *standardDurationContext) Add(d time.Duration)    { atomic.AddInt64(&ctx.tmp, int64(d)) }
func (ctx *standardDurationContext) Duration() float64 {
	var d time.Duration

	if ctx.tmp > 0 {
		d = time.Duration(ctx.tmp)
	} else if ctx.starttime.IsZero() {
		// donot start, and no duration cumulatived
		return 0
	} else {
		d = time.Since(ctx.starttime)
	}

	switch ctx.unit {
	case DurationSecond:
		return d.Seconds()
	case DurationMillisecond:
		return float64(d.Milliseconds())
	case DurationMicrosecond:
		return float64(d.Microseconds())
	case DurationNanosecond:
		return float64(d.Nanoseconds())
	}

	return 0
}

// nil duration context
type nilDurationContext struct{}

func NilDurationContext() DurationContext              { return &nilDurationContext{} }
func (ctx *nilDurationContext) SetUnit(u DurationUnit) {}
func (ctx *nilDurationContext) Start()                 {}
func (ctx *nilDurationContext) Add(d time.Duration)    {}
func (ctx *nilDurationContext) Duration() float64      { return 0 }

// standard counter context
type standardCounterContext struct {
	val   int64
	count int64
}

func NewCounterContext() CounterContext { return &standardCounterContext{} }
func (ctx *standardCounterContext) Add(v int64) {
	atomic.AddInt64(&ctx.val, v)
	atomic.AddInt64(&ctx.count, 1)
}
func (ctx *standardCounterContext) Count() int64 { return ctx.count }
func (ctx *standardCounterContext) Value() int64 { return ctx.val }

// nil counter context
type nilCounterContext struct{}

func NilCounterContext() CounterContext     { return &nilCounterContext{} }
func (ctx *nilCounterContext) Add(v int64)  {}
func (ctx *nilCounterContext) Count() int64 { return 0 }
func (ctx *nilCounterContext) Value() int64 { return 0 }

// standard duration histogram metric
type standardDurationHistogramMetric struct {
	metric prometheus.Histogram
}

func NewHistogramDurationMetric(metric prometheus.Histogram) DurationHistogram {
	return &standardDurationHistogramMetric{metric: metric}
}
func (m *standardDurationHistogramMetric) TimeAccumulator() CumulativeDurationFn {
	return func(ctx DurationContext) {
		if m.metric == nil {
			return
		}

		m.metric.Observe(ctx.Duration())
	}
}

// nil duration histogram metric
type nilDurationHistogramMetric struct{}

func NilHistogramDurationMetric() DurationHistogram {
	return &nilDurationHistogramMetric{}
}
func (m *nilDurationHistogramMetric) TimeAccumulator() CumulativeDurationFn {
	return func(dc DurationContext) {}
}

type standardTotalCountHistogram struct {
	metric prometheus.Histogram
}

func NewTotalCounterHistogram(metric prometheus.Histogram) TotalCountHistogram {
	return &standardTotalCountHistogram{metric: metric}
}
func (m *standardTotalCountHistogram) CountAccumulator() CountTotalFn {
	return func(ctx CounterContext) {
		if m.metric == nil {
			return
		}

		m.metric.Observe(float64(ctx.Value()))
	}
}

type nilTotalCountHistogram struct{}

func NilTotalCounterHistogram() TotalCountHistogram              { return &nilTotalCountHistogram{} }
func (m *nilTotalCountHistogram) CountAccumulator() CountTotalFn { return func(CounterContext) {} }
