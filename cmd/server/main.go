package main

import (
	"log"
	"net/http"
	"os"

	"github.com/hash-walker/volga-audio-transcriber/api"
	"github.com/hash-walker/volga-audio-transcriber/transcriber"
)

func main() {
	uploadDir := os.TempDir()
	resultsDir := "transcriptions"
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		log.Fatalf("failed to create results directory: %v", err)
	}

	modelSize := getEnv("WHISPER_MODEL", "base")
	device := getEnv("WHISPER_DEVICE", "cpu")
	port := getEnv("PORT", "8080")

	pipeline := transcriber.NewPipeline(modelSize, device, 2)
	server := api.NewServer(uploadDir, resultsDir, pipeline)

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	log.Printf("Starting transcription service on :%s (model=%s, device=%s)", port, modelSize, device)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
