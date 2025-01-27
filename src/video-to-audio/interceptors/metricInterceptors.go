package interceptors

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	GRPCRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_stream_requests_total",
			Help: "Total number of gRPC streaming requests",
		},
		[]string{"method", "type"}, // type: server_stream, client_stream, bidi_stream
	)

	// Histogram for stream durations
	GRPCRequestDurations = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_stream_duration_seconds",
			Help:    "Duration of gRPC streaming calls",
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
	// Register custom metrics with  Prometheus
	prometheus.MustRegister(GRPCRequestsTotal, GRPCRequestDurations, GRPCRouteLatencyGauge)
}

func PrometheusStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		// Determine the stream type
		streamType := "bidi_stream"
		if info.IsClientStream && !info.IsServerStream {
			streamType = "client_stream"
		} else if !info.IsClientStream && info.IsServerStream {
			streamType = "server_stream"
		}

		// Process the handler
		err := handler(srv, ss)

		// Record the stream duration
		statusCode := "OK"
		if err != nil {
			statusCode = status.Code(err).String()
		}
		GRPCRequestsTotal.WithLabelValues(info.FullMethod, streamType).Inc()
		duration := time.Since(start).Seconds()
		GRPCRequestDurations.WithLabelValues(info.FullMethod, streamType, statusCode).Observe(duration)
		GRPCRouteLatencyGauge.WithLabelValues(info.FullMethod, streamType, statusCode).Set(duration)

		return err
	}
}
