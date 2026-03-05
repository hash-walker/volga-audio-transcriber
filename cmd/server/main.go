package main

import (
	"log"
	"net/http"
	"os"

	"github.com/hash-walker/volga-audio-transcriber/api"
)

func main() {
	uploadDir := os.TempDir()
	//modelSize := getEnv("WHISPER_MODEL", "base")

	port := getEnv("PORT", "8080")

	server := api.NewServer(uploadDir)
	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	log.Printf("Starting transcription service on :%s", port)

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
