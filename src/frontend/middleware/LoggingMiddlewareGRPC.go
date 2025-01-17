package middleware

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryClientLoggingInterceptor logs client-side unary calls
func UnaryClientLoggingInterceptor(logger *logrus.Logger) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		startTime := time.Now()
		logger.Infof("Outgoing unary request: Method=%s | Request=%+v", method, req)

		// Make the gRPC call
		err := invoker(ctx, method, req, reply, cc, opts...)

		// Log the result
		duration := time.Since(startTime)
		if err != nil {
			grpcStatus := status.Convert(err)
			logger.Errorf("Unary call failed: Method=%s | Error=%v | GRPCStatus=%v | Duration=%v", method, err, grpcStatus.Message(), duration)
		} else {
			logger.Infof("Unary call succeeded: Method=%s | Response=%+v | Duration=%v", method, reply, duration)
		}

		return err
	}
}

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

