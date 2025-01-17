package handlers

import (
	"context"
	"io"
	"net/http"
	"net/url"

	pb "converter/frontend/genproto"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

func GetEmail(c *gin.Context) string {
	session := sessions.Default(c)
	email := session.Get("email")
	if email == nil {
		return ""
	}
	return email.(string)
}

type ImageToPdfHandler struct {
	ImageToPdfGRPCClient *pb.ImageToPdfConverterServiceClient
	*logrus.Logger
	*DBClient
}

func (s *ImageToPdfHandler) Get(c *gin.Context) {

	c.HTML(http.StatusOK, "image-to-pdf.html", gin.H{"Email": GetEmail(c)})
}

func (s *ImageToPdfHandler) Convert(c *gin.Context) {
	s.Logger.WithFields(logrus.Fields{"Method": c.Request.Method, "Path": c.Request.URL.Path}).Info("Image to PDF conversion started")

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No images uploaded"})
		return
	}

	email := GetEmail(c)
	if email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not logged in"})
		return
	}

	md := metadata.New(map[string]string{"email": email})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	stream, err := (*s.ImageToPdfGRPCClient).Convert(ctx)
	if err != nil {
		s.Logger.Errorf("Failed to create gRPC stream: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create gRPC stream"})
		return
	}

	for _, file := range files {
		f, err := file.Open()
		if err != nil {
			s.Logger.Errorf("Failed to open file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
			return
		}
		defer f.Close()

		buffer := make([]byte, 64*1024) // 64KB chunks
		for {
			n, err := f.Read(buffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				s.Logger.Errorf("Error reading file: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading file"})
				return
			}

			chunk := &pb.ConvertImageToPdfRequest{
				ImageChunk: buffer[:n],
				EndOfImage: false,
			}

			if err := stream.Send(chunk); err != nil {
				s.Logger.Errorf("Failed to send image chunk: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send image chunk"})
				return
			}
		}

		// Send an empty chunk with EndOfImage set to true to indicate the end of the image
		endChunk := &pb.ConvertImageToPdfRequest{
			EndOfImage: true,
		}
		if err := stream.Send(endChunk); err != nil {
			s.Logger.Errorf("Failed to send end of image chunk: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send end of image chunk"})
			return
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		s.Logger.Errorf("Failed to receive gRPC response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to receive gRPC response"})
		return
	}
	if err := s.DBClient.AddDownloadRecord(email, "image-to-pdf", resp.PdfPath); err != nil {
		s.Logger.Errorf("error while writing to database %v", err)

	} else {
		s.Logger.Info("successfully written to database")
	}
	c.Redirect(http.StatusFound, "/user/download?url="+url.QueryEscape(resp.PdfPath))

}
