package prometheus

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/metrics/v1"

	"github.com/google/wire"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
)

var (
	// i am bad please don't be mad
	initialized   bool
	inFlightGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "in_flight_requests",
		Help: "A gauge of requests currently being served by the wrapped handler.",
	})

	counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_total",
			Help: "A counter for requests to the wrapped handler.",
		},
		[]string{"code", "method"},
	)

	// duration is partitioned by the HTTP method and handler. It uses custom
	// buckets based on the expected request duration.
	duration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "A histogram of latencies for requests.",
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10},
		},
		[]string{"code", "method"},
	)

	// timeToWriteHeader has no labels, making it a zero-dimensional
	// ObserverVec.
	timeToWriteHeader = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "time_to_write_header",
			Help: "A histogram of latencies for requests.",
		},
		[]string{"code", "method"},
	)

	// responseSize has no labels, making it a zero-dimensional
	// ObserverVec.
	requestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "request_size_bytes",
			Help: "A histogram of response sizes for requests.",
		},
		[]string{},
	)

	// responseSize has no labels, making it a zero-dimensional
	// ObserverVec.
	responseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "response_size_bytes",
			Help: "A histogram of response sizes for requests.",
		},
		[]string{},
	)

	// Providers represents what this library offers to external users in the form of dependencies
	Providers = wire.NewSet(
		ProvideUnitCounter,
	)
)

func init() {
	if !initialized {
		for _, collector := range []prometheus.Collector{
			inFlightGauge,
			counter,
			duration,
			timeToWriteHeader,
			requestSize,
			responseSize,
		} {
			prometheus.DefaultRegisterer.Register(collector)
		}
	}
	initialized = true
}

// ProvideMiddleware is our middleware function
func ProvideMiddleware() metrics.Middleware {
	return func(next http.Handler) http.Handler {
		return promhttp.InstrumentHandlerInFlight(inFlightGauge,
			promhttp.InstrumentHandlerDuration(duration,
				promhttp.InstrumentHandlerCounter(counter,
					promhttp.InstrumentHandlerRequestSize(requestSize,
						promhttp.InstrumentHandlerResponseSize(responseSize,
							promhttp.InstrumentHandlerTimeToWriteHeader(timeToWriteHeader, next),
						),
					),
				),
			),
		)
	}
}

// ProvideMetricsHandler provides a metrics handler
func ProvideMetricsHandler() metrics.Handler {
	return promhttp.Handler()
}

// https://github.com/prometheus/client_golang/issues/412

type unitCounter struct {
	meta prometheus.Gauge
}

func (c *unitCounter) SetCount(ctx context.Context, count uint64) {
	pb := &dto.Metric{}
	c.meta.Write(pb)
	currentCount := pb.GetCounter().GetValue()
	c.meta.Add(currentCount * -1)
	c.meta.Add(float64(count))
}

func (c *unitCounter) Increment(ctx context.Context) {
	c.meta.Inc()
}

func (c *unitCounter) IncrementBy(ctx context.Context, val uint64) {
	c.meta.Add(float64(val))
}

func (c *unitCounter) Decrement(ctx context.Context) {
	c.meta.Dec()
}

// ProvideUnitCounter provides a prometheus-backed unit counter
func ProvideUnitCounter(counterName metrics.CounterName, namespace metrics.Namespace) metrics.UnitCounter {
	return &unitCounter{
		meta: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: string(namespace),
			Name:      string(counterName),
		}),
	}
}
