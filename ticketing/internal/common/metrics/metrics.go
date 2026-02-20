package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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

func (m *Metrics) MiddlewareGin(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		m.httpInFlight.Inc()
		defer m.httpInFlight.Dec()

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}

		m.httpRequests.WithLabelValues(service, method, route, status).Inc()
		m.httpDuration.WithLabelValues(service, method, route, status).Observe(time.Since(start).Seconds())
	}
}

func (m *Metrics) HandlerGin() gin.HandlerFunc {
	return gin.WrapH(promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))
}
