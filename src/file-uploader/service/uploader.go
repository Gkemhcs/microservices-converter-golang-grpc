package service

import (
	"bytes"
	"context"
	pb "converter/file-uploader/genproto"
	"encoding/json"

	"fmt"

	"io"
	"time"

	"os"

	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type FileUploaderServer struct {
	*pb.UnimplementedFileUploaderServiceServer
	*logrus.Logger
	trace.Tracer
	storageBucketHandle *storage.BucketHandle
}

func NewFileUploaderServer(bucketName string, logger *logrus.Logger, tracer trace.Tracer) (*FileUploaderServer, error) {
	// Initialize the storage client

	storageClient, err := storage.NewClient(context.Background())
	if err != nil {
		logger.Errorf("Error creating storage client: %v", err)
		return nil, err
	}

	// Try to get the bucket reference and check if it exists
	bucketHandle := storageClient.Bucket(bucketName)
	_, err = bucketHandle.Attrs(context.Background()) // Check if the bucket exists
	if err != nil {
		logger.Errorf("Error accessing bucket %s: %v", bucketName, err)
		return nil, err
	}

	return &FileUploaderServer{
		UnimplementedFileUploaderServiceServer: &pb.UnimplementedFileUploaderServiceServer{},
		Logger:                                 logger,
		storageBucketHandle:                    bucketHandle,
		Tracer:                                 tracer,
	}, nil
}

func (s *FileUploaderServer) Upload(stream pb.FileUploaderService_UploadServer) error {
	ctx := stream.Context()
	ctx, parentSpan := s.Tracer.Start(ctx, "FileUploaderServer")
	defer parentSpan.End()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.InvalidArgument, "metadata is not provided")
	}
	fileNames := md.Get("filename")
	if len(fileNames) == 0 {
		s.Logger.Error("Filename not provided in metadata")
		return status.Errorf(codes.InvalidArgument, "missing filename in metadata")
	}

	serviceType := md.Get("serviceType")
	if len(serviceType) == 0 {
		s.Logger.Error("ServiceType not provided in metadata")
		return status.Errorf(codes.InvalidArgument, "missing serviceType in metadata")
	}

	fileName := fileNames[0]
	_, childSpan := s.Tracer.Start(ctx, "Reading chunks from client stream")
	var fileBuffer bytes.Buffer

	// Read chunks from the stream.
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			s.Logger.Errorf("Error receiving chunk: %v", err)
			return status.Errorf(codes.Internal, "Error receiving chunk")
		}

		// Write the chunk content to the buffer.
		_, err = fileBuffer.Write(chunk.GetContent())
		if err != nil {
			s.Logger.Errorf("Error writing chunk to buffer: %v", err)
			return status.Errorf(codes.Internal, "Error writing chunk to buffer")
		}
	}
	childSpan.End()
	childCtx, childSpan := s.Tracer.Start(ctx, "Uploading file to GCS")
	url, err := s.uploadToGCS(childCtx, fileBuffer.Bytes(), fileName, serviceType[0])
	if err != nil {
		s.Logger.Printf("Error uploading to GCS: %v", err)
		return err
	}
	childSpan.End()
	return stream.SendAndClose(&pb.UploadResponse{Url: url})

}

func (s *FileUploaderServer) uploadToGCS(ctx context.Context, data []byte, fileName string, serviceType string) (string, error) {
	// Get the bucket and object reference.
	ctx, span := s.Tracer.Start(ctx, "uploadToGCS")
	defer span.End()
	_, childSpan := s.Tracer.Start(ctx, "File Upload To GCS")
	filename := "uploads/" + serviceType + "/" + fileName
	object := s.storageBucketHandle.Object(filename)

	// Write data to GCS.
	writer := object.NewWriter(ctx)
	_, err := writer.Write(data)
	if err != nil {
		return "", fmt.Errorf("failed to write file to GCS: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("Writer.Close: %v", err)
	}
	childSpan.End()
	_, childSpan = s.Tracer.Start(ctx, "Generate Signed URL")
	// Load the private key from the service account file
	type serviceAccountKey struct {
		PrivateKey string `json:"private_key"`
	}
	var key serviceAccountKey
	privateKeyFile, err := os.ReadFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	if err != nil {
		return "", fmt.Errorf("failed to read private key file: %w", err)
	}
	if err := json.Unmarshal(privateKeyFile, &key); err != nil {
		return "", fmt.Errorf("failed to unmarshal private key file: %w", err)
	}

	// Generate signed URL.
	signedURL, err := storage.SignedURL(s.storageBucketHandle.BucketName(), filename, &storage.SignedURLOptions{
		GoogleAccessID: os.Getenv("GCP_SERVICE_ACCOUNT"),
		PrivateKey:     []byte(key.PrivateKey),
		Method:         "GET",
		Expires:        time.Now().Add(10 * time.Minute), // Signed URL valid for 10 minutes
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}
	childSpan.End()
	return signedURL, nil
}
