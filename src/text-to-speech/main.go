package main

import (
	"converter/text-to-speech/grpcClient"
	"converter/text-to-speech/interceptors"
	"converter/text-to-speech/service"
	"converter/text-to-speech/utils"
	"fmt"
	"net/http"

	"log"
	"net"

	pb "converter/text-to-speech/genproto"


	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

)
var (
	FILE_UPLOADER_HOST=utils.GetEnv("FILE_UPLOADER_HOST","localhost")
	FILE_UPLOADER_PORT=utils.GetEnv("FILE_UPLOADER_PORT","8084")
	SERVER_HOST=utils.GetEnv("SERVER_HOST","0.0.0.0")
	SERVER_PORT=utils.GetEnv("SERVER_PORT","8081")
)




func main() {

	logger := utils.NewLogger()

	cleanup := utils.InitTracer("text-to-speech", logger)
	defer cleanup()

	tracer := otel.Tracer("text-to-speech-converter")

	uploaderClient, err := grpcClient.NewFileUploaderClient(fmt.Sprintf("%s:%s",FILE_UPLOADER_HOST,FILE_UPLOADER_PORT),logger)
	if err != nil {
		logger.Fatal("error while creating file uploader client", err)
	}
	
	pdfToTextSpeechServer := service.NewTextToSpeechServer(logger, uploaderClient, tracer)
	interceptors.InitMetrics()
	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()),
	grpc.ChainUnaryInterceptor(interceptors.UnaryLoggingInterceptor(logger),interceptors.PrometheusInterceptor()))
	
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logger.Info("server started")
		err:=http.ListenAndServe(":9090", nil)
		if err!=nil{
			logger.Panic(err)
		} // Prometheus metrics endpoint
	}()

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s",SERVER_HOST,SERVER_PORT))
	if err != nil {
		log.Fatal("error while starting to listen on port,", err)
	}
	pb.RegisterTextToSpeechConverterServiceServer(grpcServer, pdfToTextSpeechServer)

	reflection.Register(grpcServer)
	logger.Println("Server is starting on port 8081...")
	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatal("error when running,", err)
	}

}
