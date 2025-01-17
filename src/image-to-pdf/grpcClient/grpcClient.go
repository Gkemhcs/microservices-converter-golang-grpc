package grpcClient

import (
	proto "converter/image-to-pdf/genproto"
	"converter/image-to-pdf/interceptors"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	"google.golang.org/grpc/credentials/insecure"
)

func NewFileUploaderClient(address string, logger *logrus.Logger) (proto.FileUploaderServiceClient, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithChainStreamInterceptor(interceptors.ClientStreamLoggingInterceptor(logger)))
	if err != nil {
		return nil, err
	}
	return proto.NewFileUploaderServiceClient(conn), nil
}
