package interceptors

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

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

// StreamLoggingInterceptor logs streaming requests
func StreamLoggingInterceptor(logger *logrus.Logger) grpc.StreamServerInterceptor {
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
