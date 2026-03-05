package api

import (
	"net/http"
)

type Server struct {
	uploadDir string
}

func NewServer(uploadDir string) *Server {
	return &Server{
		uploadDir: uploadDir,
	}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/transcribe", s.handleTranscribe)
	mux.HandleFunc("GET /api/v1/health", s.handleHealth)
}
