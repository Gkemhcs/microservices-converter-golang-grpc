package services

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	pb "converter/image-to-pdf/genproto"

	"github.com/google/uuid"
	"github.com/signintech/gopdf"
	"github.com/sirupsen/logrus"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ImageToPdfServer struct {
	*logrus.Logger
	pb.UnimplementedImageToPdfConverterServiceServer
	UploaderClient pb.FileUploaderServiceClient
	trace.Tracer
}

func NewImageToPdfServer(logger *logrus.Logger, uploaderClient pb.FileUploaderServiceClient, tracer trace.Tracer) *ImageToPdfServer {
	return &ImageToPdfServer{
		Logger:         logger,
		UploaderClient: uploaderClient,
		Tracer:         tracer,
	}
}

func (s *ImageToPdfServer) Convert(stream pb.ImageToPdfConverterService_ConvertServer) error {

	md, ok := metadata.FromIncomingContext(stream.Context())
	useremails := md.Get("email")

	if len(useremails) == 0 {
		s.Logger.Error("useremail is not provided")
		return status.Errorf(codes.InvalidArgument, "useremail is not provided")
	}
	if !ok {
		s.Logger.Error("metadata is not provided")
		return status.Errorf(codes.InvalidArgument, "metadata is not provided")
	}

	// Create a directory in the current app folder to store the images
	tempDir := filepath.Join(".", "temp_images")
	err := os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to create temp directory: %v", err)
	}

	var imagePaths []string
	var currentImageBuffer bytes.Buffer

	// Receive all image chunks
	s.Logger.Info("Starting to receive image chunks")
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to receive image chunk: %v", err)
		}

		// Write the chunk to the current image buffer
		_, err = currentImageBuffer.Write(req.GetImageChunk())
		if err != nil {
			return status.Errorf(codes.Internal, "failed to write image chunk to buffer: %v", err)
		}

		// If end_of_image is true, save the current image buffer to a file
		if req.GetEndOfImage() {
			imagePath := filepath.Join(tempDir, fmt.Sprintf("image-%d.jpg", len(imagePaths)))
			err = os.WriteFile(imagePath, currentImageBuffer.Bytes(), 0644)
			if err != nil {
				return status.Errorf(codes.Internal, "failed to write image file: %v", err)
			}
			imagePaths = append(imagePaths, imagePath)
			currentImageBuffer.Reset()
		}
	}
	s.Logger.Info("Finished receiving image chunks")

	// Convert images to PDF
	var pdfBuffer bytes.Buffer
	err = convertImagesToPDF(imagePaths, &pdfBuffer)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to convert images to PDF: %v", err)
	}

	// Upload the PDF content to the file uploader service
	s.Logger.Info("Starting PDF upload")
	filename := fmt.Sprint(useremails[0], "-image-to-pdf-", uuid.New().String(), ".pdf")
	ctx := metadata.AppendToOutgoingContext(stream.Context(), "filename", filename+".pdf")
	ctx = metadata.AppendToOutgoingContext(ctx, "serviceType", "image-to-pdf")

	uploadStream, err := s.UploaderClient.Upload(ctx)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to create upload stream: %v", err)
	}

	// Send PDF in chunks
	buffer := make([]byte, 64*1024) // 64KB chunks
	reader := bytes.NewReader(pdfBuffer.Bytes())

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "error reading PDF buffer: %v", err)
		}

		chunk := &pb.FileChunk{
			Content: buffer[:n],
		}

		if err := uploadStream.Send(chunk); err != nil {
			return status.Errorf(codes.Internal, "failed to send PDF chunk: %v", err)
		}
	}

	// Close upload stream and get response
	resp, err := uploadStream.CloseAndRecv()
	if err != nil {
		return status.Errorf(codes.Internal, "failed to complete upload: %v", err)
	}

	s.Logger.Printf("Upload completed successfully, location: %s", resp.GetUrl())
	return stream.SendAndClose(&pb.ConvertImageToPdfResponse{
		PdfPath: resp.GetUrl(),
	})
}

func convertImagesToPDF(imagePaths []string, pdfBuffer *bytes.Buffer) error {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	pageWidth, pageHeight := gopdf.PageSizeA4.W, gopdf.PageSizeA4.H

	for _, imagePath := range imagePaths {
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			return fmt.Errorf("image file %s not found", imagePath)
		}

		pdf.AddPage()
		err := pdf.Image(imagePath, 0, 0, &gopdf.Rect{W: pageWidth, H: pageHeight})
		if err != nil {
			return fmt.Errorf("failed to add image %s to PDF: %v", imagePath, err)
		}
	}

	err := pdf.Write(pdfBuffer)
	if err != nil {
		return fmt.Errorf("failed to write PDF to buffer: %v", err)
	}

	return nil
}
