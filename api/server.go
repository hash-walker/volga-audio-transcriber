package api

import (
	"net/http"

	"github.com/hash-walker/volga-audio-transcriber/transcriber"
)

type Server struct {
	pipeline  *transcriber.Pipeline
	uploadDir string
	resultDir string
}

func NewServer(uploadDir, resultDir string, pipeline *transcriber.Pipeline) *Server {
	return &Server{
		pipeline:  pipeline,
		uploadDir: uploadDir,
		resultDir: resultDir,
	}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/transcribe", s.handleTranscribe)
	mux.HandleFunc("GET /api/v1/health", s.handleHealth)
}
