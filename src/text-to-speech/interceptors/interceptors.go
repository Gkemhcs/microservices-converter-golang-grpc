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


// UnaryLoggingInterceptor logs unary requests
func UnaryLoggingInterceptor(logger *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		startTime := time.Now()
		logger.Infof("Unary call: %s | Request: %+v", info.FullMethod, req)

		// Handle the request
		resp, err := handler(ctx, req)

		// Log result
		duration := time.Since(startTime)
		if err != nil {
			logger.Errorf("Unary call failed: %s | Error: %v | Duration: %v", info.FullMethod, err, duration)
		} else {
			logger.Infof("Unary call succeeded: %s | Response: %+v | Duration: %v", info.FullMethod, resp, duration)
		}

		return resp, err
	}
}
