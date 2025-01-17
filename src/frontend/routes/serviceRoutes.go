package routes

import (
	proto "converter/frontend/genproto"
	"converter/frontend/grpcClient"
	"converter/frontend/handlers"
	"converter/frontend/middleware"
	"converter/frontend/utils"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"go.opentelemetry.io/otel/trace"
)

var (
	TEXT_TO_SPEECH__HOST = utils.GetEnv("TEXT_TO_SPEECH_HOST", "localhost")
	TEXT_TO_SPEECH__PORT = utils.GetEnv("TEXT_TO_SPEECH_PORT", "8081")
	VIDEO_TO_AUDIO_HOST  = utils.GetEnv("VIDEO_TO_AUDIO_HOST", "localhost")
	VIDEO_TO_AUDIO_PORT  = utils.GetEnv("VIDEO_TO_AUDIO_PORT", "8082")
	IMAGE_TO_PDF_HOST    = utils.GetEnv("IMAGE_TO_PDF_HOST", "localhost")
	IMAGE_TO_PDF_PORT    = utils.GetEnv("IMAGE_TO_PDF_PORT", "8083")
)

func SetupServiceRoutes(r *gin.Engine, logger *logrus.Logger, tracer *trace.Tracer, dbClient *handlers.DBClient) {
	serviceRoutes := r.Group("/services")

	grpcClients := grpcClient.GRPCClients{
		TextToSpeechClient: proto.NewTextToSpeechConverterServiceClient(
			grpcClient.DialGRPCServer(fmt.Sprintf("%s:%s", TEXT_TO_SPEECH__HOST, TEXT_TO_SPEECH__PORT), logger)),
		VideoToAudioClient: proto.NewVideoToAudioConverterServiceClient(grpcClient.DialGRPCServer(fmt.Sprintf("%s:%s", VIDEO_TO_AUDIO_HOST, VIDEO_TO_AUDIO_PORT), logger)),
		ImageToPdfClient:   proto.NewImageToPdfConverterServiceClient(grpcClient.DialGRPCServer(fmt.Sprintf("%s:%s", IMAGE_TO_PDF_HOST, IMAGE_TO_PDF_PORT), logger)),
	}
	pdfToDocxHandler := handlers.ImageToPdfHandler{
		Logger:               logger,
		ImageToPdfGRPCClient: &grpcClients.ImageToPdfClient,
		DBClient:             dbClient}
	textToSpeechHandler := handlers.TextToSpeechHandler{
		Logger:                 logger,
		TextToSpeechGRPCClient: &grpcClients.TextToSpeechClient,
		Tracer:                 *tracer,
		DBClient:               dbClient,
	}
	videoToAudioHandler := handlers.VideoToAudioHandler{
		Logger:                 logger,
		VideoToAudioGRPCClient: &grpcClients.VideoToAudioClient,
		Tracer:                 *tracer,
		DBClient:               dbClient,
	}

	serviceRoutes.Use(middleware.AuthMiddleware())
	{
		serviceRoutes.GET("/image-to-pdf", pdfToDocxHandler.Get)
		serviceRoutes.POST("/image-to-pdf/convert", pdfToDocxHandler.Convert)
		serviceRoutes.GET("/text-to-speech", textToSpeechHandler.Get)
		serviceRoutes.POST("/text-to-speech/convert", textToSpeechHandler.Convert)
		serviceRoutes.GET("/video-to-audio", videoToAudioHandler.Get)
		serviceRoutes.POST("/video-to-audio/convert", videoToAudioHandler.Convert)
	}
}
