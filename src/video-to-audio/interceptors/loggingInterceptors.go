package interceptors

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// ClientStreamLoggingInterceptor logs client-streaming calls
func ClientStreamLoggingInterceptor(logger *logrus.Logger) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		startTime := time.Now()
		logger.Infof("Client stream call started: %s | IsClientStream: %v | IsServerStream: %v", method, desc.ClientStreams, desc.ServerStreams)

		// Start the client stream
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			logger.Errorf("Client stream call failed: %s | Error: %v", method, err)
		} else {
			logger.Infof("Client stream call established: %s | Duration: %v", method, time.Since(startTime))
		}

		return clientStream, err
	}
}

// StreamLoggingInterceptor logs streaming requests
func ServerStreamLoggingInterceptor(logger *logrus.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		startTime := time.Now()
		logger.Infof("Stream call started: %s | IsClientStream: %v | IsServerStream: %v", info.FullMethod, info.IsClientStream, info.IsServerStream)

		// Handle the stream
		err := handler(srv, ss)

		// Log result
		duration := time.Since(startTime)
		if err != nil {
			logger.Errorf("Stream call failed: %s | Error: %v | Duration: %v", info.FullMethod, err, duration)
		} else {
			logger.Infof("Stream call succeeded: %s | Duration: %v", info.FullMethod, duration)
		}

		return err
	}
}
