package service

import (
	pb "converter/video-to-audio/genproto"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go.opentelemetry.io/otel/trace"
)

type VideoToAudioServer struct {
	*logrus.Logger
	pb.UnimplementedVideoToAudioConverterServiceServer
	UploaderClient pb.FileUploaderServiceClient
	trace.Tracer
}

func NewVideoToAudioServer(logger *logrus.Logger, uploaderClient pb.FileUploaderServiceClient, tracer trace.Tracer) *VideoToAudioServer {
	return &VideoToAudioServer{
		Logger:         logger,
		UploaderClient: uploaderClient,
		Tracer:         tracer,
	}
}

func (s *VideoToAudioServer) Convert(stream pb.VideoToAudioConverterService_ConvertServer) error {
	ctx, span := s.Tracer.Start(stream.Context(), "Video to Audio Conversion")
	defer span.End()

	// Create buffer to store incoming video
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.InvalidArgument, "metadata is not provided")
	}
	userEmails := md.Get("email")
	if len(userEmails) == 0 {
		s.Logger.Error("Email not provided in metadata")
		return status.Errorf(codes.InvalidArgument, "missing email in metadata")
	}
	filename := fmt.Sprint(userEmails[0], "-video-to-audio-", uuid.New().String())

	// Create pipe for FFmpeg input and output
	videoReader, videoWriter := io.Pipe()
	audioReader, audioWriter := io.Pipe()
	var conversionErr error

	// Start FFmpeg conversion
	s.Logger.Info("Starting audio conversion")
	go func() {
		defer audioWriter.Close()

		_, conversionSpan := s.Tracer.Start(ctx, "Audio Conversion")
		defer conversionSpan.End()

		err := ffmpeg.Input("pipe:").
			Output("pipe:", ffmpeg.KwArgs{
				"acodec": "libmp3lame",
				"f":      "mp3",
				"vn":     "",
				"ab":     "192k",
			}).
			WithInput(videoReader).
			WithOutput(audioWriter).
			Run()

		if err != nil {
			s.Logger.Errorf("FFmpeg conversion error: %v", err)
			conversionErr = err
			audioWriter.CloseWithError(err) // Ensure error is properly handled
			return
		}
		s.Logger.Info("Conversion completed successfully")
	}()

	// Receive video chunks and write to videoWriter
	go func() {
		defer videoWriter.Close()
		_, receiveSpan := s.Tracer.Start(ctx, "Receiving Video Chunks")
		defer receiveSpan.End()

		for {
			chunk, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				s.Logger.Errorf("Failed to receive video chunk: %v", err)
				videoWriter.CloseWithError(err)
				return
			}

			_, err = videoWriter.Write(chunk.GetChunk())
			if err != nil {
				s.Logger.Errorf("Failed to write chunk to video writer: %v", err)
				videoWriter.CloseWithError(err)
				return
			}
		}
	}()

	// Start upload after conversion is complete
	s.Logger.Info("Starting audio upload")
	uploadCtx, uploadSpan := s.Tracer.Start(ctx, "Uploading Audio Chunks")
	defer uploadSpan.End()
	uploadCtx = metadata.AppendToOutgoingContext(uploadCtx, "filename", filename+".mp3")
	uploadCtx = metadata.AppendToOutgoingContext(uploadCtx, "serviceType", "video-to-audio")

	uploadStream, err := s.UploaderClient.Upload(uploadCtx)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to create upload stream: %v", err)
	}

	// Send audio in chunks
	buffer := make([]byte, 64*1024) // 64KB chunks

	for {
		n, err := audioReader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "error reading audio buffer: %v", err)
		}

		chunk := &pb.FileChunk{
			Content: buffer[:n],
		}

		if err := uploadStream.Send(chunk); err != nil {
			return status.Errorf(codes.Internal, "failed to send audio chunk: %v", err)
		}
	}

	if conversionErr != nil {
		return status.Errorf(codes.Internal, "conversion failed: %v", conversionErr)
	}

	// Close upload stream and get response
	resp, err := uploadStream.CloseAndRecv()
	if err != nil {
		return status.Errorf(codes.Internal, "failed to complete upload: %v", err)
	}

	s.Logger.Printf("Upload completed successfully, location: %s", resp.GetUrl())
	return stream.SendAndClose(&pb.ConvertVideoToAudioResponse{
		Url: resp.GetUrl(),
	})
}
