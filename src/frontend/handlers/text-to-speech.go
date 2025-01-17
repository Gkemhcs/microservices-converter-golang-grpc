package handlers

import (
	proto "converter/frontend/genproto"

	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/otel/trace"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

type TextToSpeechHandler struct {
	TextToSpeechGRPCClient *proto.TextToSpeechConverterServiceClient
	*logrus.Logger
	trace.Tracer
	*DBClient
}

func (s *TextToSpeechHandler) Get(c *gin.Context) {
	c.HTML(http.StatusOK, "text-to-speech.html", gin.H{"Email": GetEmail(c)})

}
func (s *TextToSpeechHandler) Convert(c *gin.Context) {


	ctx, parentSpan := s.Tracer.Start(c.Request.Context(), "Frontend Text-To-Speech Convert")
	defer parentSpan.End()
	email := GetEmail(c)

	resp, err := (*s.TextToSpeechGRPCClient).Convert(
		metadata.AppendToOutgoingContext(ctx, "userEmail", email),
		&proto.ConvertTextToSpeechRequest{Text: c.Request.FormValue("textInput")})
	if err != nil {
		s.Logger.WithFields(logrus.Fields{"Method": c.Request.Method, "Endpoint": c.Request.URL.Path}).Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error in Text to Speech conversion"})
		return
	}

	
	s.DBClient.AddDownloadRecord(email,"text-to-speech",resp.GetUrl())
	c.Redirect(http.StatusFound, "/user/download?url="+url.QueryEscape(resp.GetUrl()))
}
