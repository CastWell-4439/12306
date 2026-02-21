package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	registry     *prometheus.Registry
	httpRequests *prometheus.CounterVec
	httpDuration *prometheus.HistogramVec
	httpInFlight prometheus.Gauge
}

func New(service string) *Metrics {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	requests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests by service, method, route and status.",
		},
		[]string{"service", "method", "route", "status"},
	)
	duration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "route", "status"},
	)
	inFlight := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "http_in_flight_requests",
			Help:        "In-flight HTTP requests.",
			ConstLabels: prometheus.Labels{"service": service},
		},
	)
	reg.MustRegister(requests, duration, inFlight)

	return &Metrics{
		registry:     reg,
		httpRequests: requests,
		httpDuration: duration,
		httpInFlight: inFlight,
	}
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Middleware returns a go-zero compatible middleware for metrics collection.
func (m *Metrics) Middleware(service string) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			m.httpInFlight.Inc()
			defer m.httpInFlight.Dec()

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next(rw, r)

			status := strconv.Itoa(rw.statusCode)
			method := r.Method
			route := r.URL.Path

			m.httpRequests.WithLabelValues(service, method, route, status).Inc()
			m.httpDuration.WithLabelValues(service, method, route, status).Observe(time.Since(start).Seconds())
		}
	}
}

// Handler returns an http.HandlerFunc that serves prometheus metrics.
func (m *Metrics) Handler() http.HandlerFunc {
	h := promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}


