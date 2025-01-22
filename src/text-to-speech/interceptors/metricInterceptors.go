package interceptors

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	// Counter for gRPC requests
	GRPCRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "type", "status"},
	)

	// Histogram for gRPC request durations
	GRPCRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Histogram of gRPC request durations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "type", "status"},
	)

	// Gauge for active gRPC requests
	GRPCRouteLatencyGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "grpc_method_latency_seconds",
			Help: "Gauge of the current latency (seconds) for each method",
		},
		[]string{"method", "type", "status"},
	)
)

func InitMetrics() {
	// Register custom metrics with Prometheus
	prometheus.MustRegister(GRPCRequestsTotal, GRPCRequestDuration, GRPCRouteLatencyGauge)
}

func PrometheusInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Handle the request
		resp, err := handler(ctx, req)

		// Record the request
		statusCode := "OK"
		if err != nil {
			statusCode = status.Code(err).String()
		}
		duration := time.Since(start).Seconds()
		GRPCRequestsTotal.WithLabelValues(info.FullMethod, "Unary", statusCode).Inc()
		GRPCRequestDuration.WithLabelValues(info.FullMethod, "Unary", statusCode).Observe(duration)
		GRPCRouteLatencyGauge.WithLabelValues(info.FullMethod, "Unary", statusCode).Set(duration)
		return resp, err
	}
}
