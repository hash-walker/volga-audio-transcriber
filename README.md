# Volga Audio Transcriber

built in Go, leveraging the OpenAI's Whisper model. It features a parallel processing pipeline, automatic audio normalization, and human-readable timestamping

---

## Getting Started

### Prerequisites
- **Go** (1.21 or later)
- **Python 3.8+**
- **FFmpeg** (Required for audio normalization and chunking)

### Installation
1.  **Clone the repository:**
    ```bash
    git clone https://github.com/hash-walker/volga-audio-transcriber.git
    cd volga-audio-transcriber
    ```

2.  **Set up the Python Virtual Environment:**
    ```bash
    python3 -m venv venv
    source venv/bin/activate
    pip install openai-whisper
    ```
    *Note: The Go service expects the `whisper` CLI to be available at `venv/bin/whisper`.*

3.  **Install Go dependencies:**
    ```bash
    go mod download
    ```

### Running the Server
```bash
go run cmd/server/main.go
```
The server will start on `http://localhost:8080`.

### Usage Example
Transcribe an audio file using `curl`:
```bash
curl -X POST -F "audio=@/path/to/your/audio.mp3" http://localhost:8080/transcribe
```

---

## Architectural Design Decisions

### 1. The "Clean CLI" Pattern
Instead of wrestling with complex CGO or Python-in-Go bindings, we use a decoupled execution pattern. The Go service manages a local Python environment and executes the `whisper` command directly as a subprocess. This keeps the backend lightweight, easy to debug, and avoids memory leaks associated with embedding Python.

### 2. Parallel "Chunking" Pipeline
Processing large files (e.g., 1-hour podcasts) is memory-intensive. Our pipeline:
1.  **Splits** the file into 30-second chunks.
2.  Adds a **2-second overlap** so words aren't cut in half at boundaries.
3.  Processes chunks **concurrently** using Go's worker pool (Goroutines).
4.  **Merges** and deduplicates segments into a final, continuous transcript.

### 3. Automatic Normalization
AI models require consistent input. We use **FFmpeg** as a middleware layer to automatically normalize any uploaded format (MP3, M4A, FLAC) into a "Golden Standard": **16kHz, Mono, Signed-16Bit WAV** before processing.

---

## Technical Implementation Details

### Accepting Audio Files
We use a standard RESTful API in Go (`api/handlers.go`). Every upload is assigned a unique UUID and stored in its own session directory to ensure isolation.

### Human-Readable Timestamps
Whisper provides raw timing, but we implement a formatting layer (`transcriber/engine.go`) that converts durations into clear `HH:MM:SS` format, making the results ready for user-facing UIs.

---

## Production Scalability 

### Handling High Traffic
- **Worker Clusters:** Run a fleet of dedicated transcription workers (with GPU acceleration) that pull jobs from the queue using postgres as queue for time being.

### Storage & Resilience
- **Object Storage:** Store raw audio S3 and transcription as meta data in postgres with path of the audio 
- **Job States:** Implement a proper state machine (`PENDING`, `PROCESSING`, `SUCCESS`, `FAILED`) in a database like PostgreSQL.
- **Retries:** Use "Visibility Timeouts" in the queue to automatically retry jobs if a worker node crashes mid-task.

---

**Git Repository:** [https://github.com/hash-walker/volga-audio-transcriber](https://github.com/hash-walker/volga-audio-transcriber)