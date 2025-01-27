package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number  of HTTP requests",
		},
		[]string{"method", "route", "status"},
	)

	// Histogram metric
	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of response latency (seconds) per route",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route", "status"},
	)
	RouteLatencyGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_route_latency_seconds",
			Help: "Gauge of the current latency (seconds) for each route",
		},
		[]string{"method", "route"},
	)
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process the request
		c.Next()

		// Calculate latency

		status := c.Writer.Status()
		route := c.FullPath() // Route, e.g., "/api/v1/resource"
		if route == "" {
			route = "unknown" // Handle unmatched routes
		}

		//Counter Metric
		HttpRequestsTotal.WithLabelValues(c.Request.Method, route, http.StatusText(status)).Inc()
		duration := time.Since(start).Seconds()

		//Histogram Metric
		HttpRequestDuration.WithLabelValues(c.Request.Method, route, http.StatusText(status)).Observe(duration)

		// Gauge Metric
		RouteLatencyGauge.WithLabelValues(c.Request.Method, route).Set(duration)
	}
}
