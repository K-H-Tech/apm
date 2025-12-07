package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Tracks the number of HTTP requests.",
		}, []string{"url", "method", "code"},
	)
	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Tracks the latencies for HTTP requests.",
			Buckets: prometheus.ExponentialBuckets(0.1, 1.5, 5),
		},
		[]string{"url", "method", "code"},
	)
	requestSize = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_request_size_bytes",
			Help: "Tracks the size of HTTP requests.",
		},
		[]string{"url", "method", "code"},
	)
	responseSize = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_response_size_bytes",
			Help: "Tracks the size of HTTP responses.",
		},
		[]string{"url", "method", "code"},
	)
)

// HTTPMetrics will apply metrics for http requests on top on gin
func (m *Middlewares) HTTPMetrics() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		r := ctx.Request
		w := ctx.Writer

		// ignore /metrics endpoint
		if r.URL.Path == "/metrics" {
			ctx.Next()
			return
		}

		// set the start time for the duration observe
		startTime := time.Now()

		// execute normal process.
		ctx.Next()

		labels := []string{ctx.FullPath(), r.Method, strconv.Itoa(w.Status())}

		// set requests total
		requestsTotal.WithLabelValues(labels...).Inc()

		// set request body size
		// since r.ContentLength can be negative (in some occasions) guard the operation
		if r.ContentLength >= 0 {
			requestSize.WithLabelValues(labels...).Observe(float64(r.ContentLength))
		}

		// set request duration
		requestDuration.WithLabelValues(labels...).Observe(time.Since(startTime).Seconds())

		// set response size
		if w.Size() > 0 {
			responseSize.WithLabelValues(labels...).Observe(float64(w.Size()))
		}
	}
}
