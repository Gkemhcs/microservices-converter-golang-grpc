package main

import (
	pb "converter/file-uploader/genproto"
	"converter/file-uploader/interceptors"
	"converter/file-uploader/service"
	"converter/file-uploader/utils"
	"net/http"

	"fmt"
	"net"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	SERVER_HOST = utils.GetEnv("SERVER_HOST", "0.0.0.0")
	SERVER_PORT = utils.GetEnv("SERVER_PORT", "8081")
)

func main() {
	// Initialize logger
	logger := utils.NewLogger()
	cleanup := utils.InitTracer("file-uploader", logger)
	defer cleanup()
	tracer := otel.Tracer("file-uploader")
	logger.Info("Starting FileUploaderService...")

	// Create FileUploaderServer
	server, err := service.NewFileUploaderServer(os.Getenv("GCS_BUCKET_NAME"), logger, tracer)
	if err != nil {
		logger.Fatalf("Error creating FileUploaderServer: %v", err)
	}
	interceptors.InitMetrics()

	// Create gRPC server
	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(interceptors.UnaryLoggingInterceptor(logger)),
		grpc.ChainStreamInterceptor(interceptors.StreamLoggingInterceptor(logger), interceptors.PrometheusStreamInterceptor()))
	pb.RegisterFileUploaderServiceServer(grpcServer, server)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logger.Info("metrics server started")
		err := http.ListenAndServe(":9090", nil)
		if err != nil {
			logger.Panic(err)
		} // Prometheus metrics endpoint
	}()
	// Start listening
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", SERVER_HOST, SERVER_PORT))
	if err != nil {
		logger.Fatalf("Error while starting to listen on port: %v", err)
	}
	reflection.Register(grpcServer)

	// Start gRPC server
	logger.Info("Server is starting on port 8084...")
	err = grpcServer.Serve(listener)
	if err != nil {
		logger.Fatalf("Error when running server: %v", err)
	}
}
