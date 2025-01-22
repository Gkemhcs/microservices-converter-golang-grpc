package main

import (
	pb "converter/video-to-audio/genproto"
	"converter/video-to-audio/grpcClient"
	"converter/video-to-audio/interceptors"
	"converter/video-to-audio/service"
	"converter/video-to-audio/utils"
	"fmt"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	FILE_UPLOADER_HOST = utils.GetEnv("FILE_UPLOADER_HOST", "localhost")
	FILE_UPLOADER_PORT = utils.GetEnv("FILE_UPLOADER_PORT", "8084")
	SERVER_HOST        = utils.GetEnv("SERVER_HOST", "0.0.0.0")
	SERVER_PORT        = utils.GetEnv("SERVER_PORT", "8082")
)

func main() {
	logger := utils.NewLogger()
	cleanup := utils.InitTracer("video-to-audio", logger)
	defer cleanup()
	uploaderClient, err := grpcClient.NewFileUploaderClient(fmt.Sprintf("%s:%s", FILE_UPLOADER_HOST, FILE_UPLOADER_PORT), logger)
	if err != nil {
		logger.Fatal("error while creating file uploader client", err)
	}
	tracer := otel.Tracer("video-to-audio-converter")
	server := service.NewVideoToAudioServer(logger, uploaderClient, tracer)

	interceptors.InitMetrics()

	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainStreamInterceptor(interceptors.ServerStreamLoggingInterceptor(logger), interceptors.PrometheusStreamInterceptor()))

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logger.Info("metrics server started")
		err := http.ListenAndServe(":9090", nil)
		if err != nil {
			logger.Panic(err)
		} // Prometheus metrics endpoint
	}()
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", SERVER_HOST, SERVER_PORT))
	if err != nil {
		logger.Fatal("error while starting to listen on port,", err)
	}
	logger.Println("Server is starting on port 8082...")

	pb.RegisterVideoToAudioConverterServiceServer(grpcServer, server)
	reflection.Register(grpcServer)
	err = grpcServer.Serve(listener)

	if err != nil {
		logger.Fatal("error when running,", err)
	}
}
