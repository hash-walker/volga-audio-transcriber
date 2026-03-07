# =============================================================================
# Stage 1: Build the Go binary
# =============================================================================
FROM golang:1.25-bookworm AS builder

WORKDIR /app

# Cache dependency downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /transcriber ./cmd/server

# =============================================================================
# Stage 2: Runtime with ffmpeg + whisper
# =============================================================================
FROM python:3.11-slim-bookworm

# Install ffmpeg (for audio normalization & chunking)
RUN apt-get update && \
    apt-get install -y --no-install-recommends ffmpeg && \
    rm -rf /var/lib/apt/lists/*

# Install OpenAI Whisper CLI
RUN pip install --no-cache-dir openai-whisper

# Create whisper output directory
RUN mkdir -p /tmp/whisper_out

# Copy the Go binary from builder
COPY --from=builder /transcriber /usr/local/bin/transcriber

# Environment defaults
ENV PORT=8080
ENV WHISPER_MODEL=base
ENV WHISPER_DEVICE=cpu

EXPOSE 8080

ENTRYPOINT ["transcriber"]
