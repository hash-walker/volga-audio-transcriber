package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/hash-walker/volga-audio-transcriber/transcriber"
)

func (s *Server) handleTranscribe(w http.ResponseWriter, r *http.Request) {
	// Limit upload size to 500MB
	r.Body = http.MaxBytesReader(w, r.Body, 500<<20)

	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "missing or invalid audio file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Generate unique session ID and directory
	sessionID := uuid.New().String()
	sessionDir := filepath.Join(s.resultDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		http.Error(w, "failed to create session directory", http.StatusInternalServerError)
		return
	}

	// Save uploaded file to the session directory
	ext := filepath.Ext(header.Filename)
	audioFilename := "audio" + ext
	savePath := filepath.Join(sessionDir, audioFilename)
	dst, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}
	dst.Close()

	// Optional parameters from query/form
	language := r.FormValue("language")
	model := r.FormValue("model")

	req := transcriber.TranscriptionRequest{
		FilePath: savePath,
		Language: language,
		Model:    model,
	}

	result, err := s.pipeline.Run(r.Context(), req)
	if err != nil {
		// Clean up the directory if transcription fails
		os.RemoveAll(sessionDir)
		http.Error(w, fmt.Sprintf("transcription failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Use our newly generated ID
	result.ID = sessionID

	// Save result.json
	resultPath := filepath.Join(sessionDir, "result.json")
	resultFile, _ := os.Create(resultPath)
	defer resultFile.Close()
	json.NewEncoder(resultFile).Encode(result)

	// Save transcript.txt
	transcriptPath := filepath.Join(sessionDir, "transcript.txt")
	os.WriteFile(transcriptPath, []byte(result.FullText), 0644)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
