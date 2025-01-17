package main

import (
	"fmt"

	"net"

	pb "converter/image-to-pdf/genproto"
	"converter/image-to-pdf/grpcClient"
	"converter/image-to-pdf/interceptors"
	"converter/image-to-pdf/services"
	"converter/image-to-pdf/utils"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
)

var (
	FILE_UPLOADER_HOST = utils.GetEnv("FILE_UPLOADER_HOST", "localhost")
	FILE_UPLOADER_PORT = utils.GetEnv("FILE_UPLOADER_PORT", "8084")
	SERVER_HOST        = utils.GetEnv("SERVER_HOST", "0.0.0.0")
	SERVER_PORT        = utils.GetEnv("SERVER_PORT", "8083")
)

func main() {
	logger := utils.NewLogger()
	cleanup := utils.InitTracer("image-to-pdf", logger)
	defer cleanup()
	tracer := otel.Tracer("image-topdf-converter")
	uploaderClient, err := grpcClient.NewFileUploaderClient(fmt.Sprintf("%s:%s", FILE_UPLOADER_HOST, FILE_UPLOADER_PORT), logger)
	if err != nil {
		logger.Panic("failed to start uploaderClient:-", err)
	}
	server := services.NewImageToPdfServer(logger, uploaderClient, tracer)
	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainStreamInterceptor(interceptors.ServerStreamLoggingInterceptor(logger)))
		
	pb.RegisterImageToPdfConverterServiceServer(grpcServer, server)

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", SERVER_HOST, SERVER_PORT))
	if err != nil {
		logger.Fatalf("Failed to listen on port 8083: %v", err)
	}

	logger.Info("Starting gRPC server on port 8083")
	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
