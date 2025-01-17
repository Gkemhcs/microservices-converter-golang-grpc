package handlers

import (
	"bytes"

	pb "converter/frontend/genproto"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

type VideoToAudioHandler struct {
	VideoToAudioGRPCClient *pb.VideoToAudioConverterServiceClient
	*logrus.Logger
	trace.Tracer
	*DBClient
}

func (s *VideoToAudioHandler) Get(c *gin.Context) {
	c.HTML(http.StatusOK, "video-to-audio.html", gin.H{"Email": GetEmail(c)})
}

func (s *VideoToAudioHandler) Convert(c *gin.Context) {

	ctx, span := s.Tracer.Start(c.Request.Context(), "Frontend Video To Audio Converter Route")
	defer span.End()
	videoFile, headers, err := c.Request.FormFile("videoFile")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse uploaded file"})
		return
	}
	s.Logger.Info("file size", headers.Size, headers.Filename)
	defer videoFile.Close()

	// Create a buffer to store the video file
	var videoBuffer bytes.Buffer
	_, err = io.Copy(&videoBuffer, videoFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read video file"})
		return
	}

	// Get user email from session
	email := GetEmail(c)
	if email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not logged in"})
		return
	}

	// Create gRPC metadata with user email
	md := metadata.New(map[string]string{"email": email})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Create gRPC stream
	stream, err := (*s.VideoToAudioGRPCClient).Convert(ctx)
	if err != nil {
		s.Logger.Errorf("Failed to create gRPC stream: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create gRPC stream"})
		return
	}

	// Stream video file in chunks
	buffer := make([]byte, 64*1024) // 64KB chunks
	reader := bytes.NewReader(videoBuffer.Bytes())

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			s.Logger.Errorf("Error reading video buffer: %v", err)
			stream.CloseSend() // Ensure stream is properly closed
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading video buffer"})
			return
		}

		chunk := &pb.ConvertVideoToAudioRequest{
			Chunk: buffer[:n],
		}

		if err := stream.Send(chunk); err != nil {
			s.Logger.Errorf("Failed to send video chunk: %v", err)
			stream.CloseSend() // Ensure stream is properly closed
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send video chunk"})
			return
		}
	}

	// Close the stream and receive the response
	resp, err := stream.CloseAndRecv()
	if err != nil {
		s.Logger.Errorf("Failed to receive gRPC response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to receive gRPC response"})
		return
	}
	s.DBClient.AddDownloadRecord(email, "video-to-audio", resp.GetUrl())
	c.Redirect(http.StatusFound, "/user/download?url="+url.QueryEscape(resp.GetUrl()))

}
