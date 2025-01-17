package service

import (
	"context"
	pb "converter/text-to-speech/genproto"
	"fmt"
	"io"
	"os"

	"github.com/Duckduckgot/gtts"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc/status"
)

const (
	chunkSize = 1024 * 1024
)

type TextToSpeechServer struct {
	*gtts.Speech
	*logrus.Logger
	Tracer trace.Tracer
	pb.UnimplementedTextToSpeechConverterServiceServer
	UploaderClient pb.FileUploaderServiceClient
}

func NewTextToSpeechServer(logger *logrus.Logger, uploaderClient pb.FileUploaderServiceClient, tracer trace.Tracer) *TextToSpeechServer {
	return &TextToSpeechServer{
		Speech: &gtts.Speech{
			Folder:   "output",
			Language: "en",
		},
		Logger:         logger,
		UploaderClient: uploaderClient,
		Tracer:         tracer,
	}
}

func (server *TextToSpeechServer) Convert(ctx context.Context, req *pb.ConvertTextToSpeechRequest) (*pb.ConvertTextToSpeechResponse, error) {

	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		server.Logger.Error("metadata is not provided")

		return nil, status.Errorf(codes.InvalidArgument, "metadata is not provided")
	}

	useremails := md.Get("userEmail")

	if len(useremails) == 0 {
		server.Logger.Error("useremail is not provided")
		return nil, status.Errorf(codes.InvalidArgument, "useremail is not provided")
	}

	server.Logger.Info("file conversion started")

	text := req.GetText()
	filename := fmt.Sprint(useremails[0], "-text-to-speech-", uuid.New().String())
	outputfile, err := server.Speech.CreateSpeechFile(text, filename)

	if err != nil {
		server.Logger.Error("can't convert file")
		return nil, status.Errorf(codes.Internal, "can't convert file")
	}
	file, err := os.Open(outputfile)
	if err != nil {
		server.Logger.Fatalf("Failed to open file: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to open file")
	}
	defer file.Close()
	defer os.Remove(outputfile)
	ctx = metadata.AppendToOutgoingContext(ctx, "filename", filename+".mp3")
	ctx = metadata.AppendToOutgoingContext(ctx, "serviceType", "text-to-speech")
	stream, err := server.UploaderClient.Upload(ctx)
	if err != nil {
		server.Logger.Fatalf("Failed to start stream: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to start stream")
	}

	buffer := make([]byte, chunkSize)
	for {
		// Read a chunk of data from the file
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			server.Logger.Fatalf("Error reading file: %v", err)
			return nil, status.Errorf(codes.Internal, "Error reading file")
		}

		// If we've reached the end of the file, close the stream
		if n == 0 {
			break
		}

		// Create a new FileChunk message
		fileChunk := &pb.FileChunk{
			Content: buffer[:n], // Send the chunk to the server
		}
		if err := stream.Send(fileChunk); err != nil {
			server.Logger.Fatalf("Error sending chunk: %v", err)
			return nil, status.Errorf(codes.Internal, "Error sending chunk")
		}

		server.Logger.Printf("Sent chunk of size: %d bytes\n", n)

	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		server.Logger.Errorf("Error closing stream: %v", err)
		return nil, status.Errorf(codes.Internal, "Error closing stream")
	}

	return &pb.ConvertTextToSpeechResponse{Url: resp.Url}, nil
}
