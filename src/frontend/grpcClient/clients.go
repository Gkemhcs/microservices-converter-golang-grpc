package grpcClient

import (
	"context"
	"log"

	proto "converter/frontend/genproto"
	"converter/frontend/middleware"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClients struct {
	TextToSpeechClient proto.TextToSpeechConverterServiceClient
	VideoToAudioClient proto.VideoToAudioConverterServiceClient
	ImageToPdfClient   proto.ImageToPdfConverterServiceClient
}

func DialGRPCServer(address string,logger *logrus.Logger) *grpc.ClientConn {

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()),
	 grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	 grpc.WithChainUnaryInterceptor(middleware.UnaryClientLoggingInterceptor(logger)),
	grpc.WithChainStreamInterceptor(middleware.ClientStreamLoggingInterceptor(logger)))

	if err != nil {
		log.Fatalf("Failed to connect to GRPC server: %v", err)
	}

	return conn

}

// Define the service clients here
type TextToSpeechServiceClient interface {
	Convert(ctx context.Context, in *proto.ConvertTextToSpeechRequest, opts ...grpc.CallOption) (*proto.ConvertTextToSpeechResponse, error)
}

type VideoToAudioServiceClient interface {
	Convert(ctx context.Context, in *proto.ConvertTextToSpeechRequest, opts ...grpc.CallOption) (*proto.ConvertTextToSpeechResponse, error)
}

type PDFToDocxServiceClient interface {
	Convert(ctx context.Context, in *proto.ConvertTextToSpeechRequest, opts ...grpc.CallOption) (*proto.ConvertTextToSpeechResponse, error)
}
