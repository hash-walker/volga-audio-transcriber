package transcriber

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// WhisperEngine wraps OpenAI's Whisper CLI (whisper or faster-whisper).
// Using CLI over Python bindings keeps our Go service clean
// and avoids CGO/Python interop complexity.
type WhisperEngine struct {
	ModelSize string // "tiny", "base", "small", "medium", "large"
	Device    string // "cpu" or "cuda"
}

// whisperSegment matches Whisper's JSON output format.
type whisperSegment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

type whisperOutput struct {
	Segments []whisperSegment `json:"segments"`
	Text     string           `json:"text"`
	Language string           `json:"language"`
}

func NewWhisperEngine(modelSize, device string) *WhisperEngine {
	if modelSize == "" {
		modelSize = "base"
	}
	if device == "" {
		device = "cpu"
	}
	return &WhisperEngine{ModelSize: modelSize, Device: device}
}

// TranscribeChunk transcribes a single audio chunk and returns segments.
func (w *WhisperEngine) TranscribeChunk(chunkPath string, offsetDuration time.Duration) ([]Segment, error) {
	cwd, _ := os.Getwd()
	whisperExe := filepath.Join(cwd, "venv", "bin", "whisper")
	outputDir := filepath.Dir(chunkPath)

	// Execute the whisper CLI directly
	fmt.Printf("Starting transcription for %s (model: %s)...\n", chunkPath, w.ModelSize)
	start := time.Now()
	cmd := exec.Command(whisperExe, chunkPath,
		"--model", w.ModelSize,
		"--device", w.Device,
		"--output_format", "json",
		"--output_dir", outputDir,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("transcription failed: %w\nOutput: %s", err, string(output))
	}
	fmt.Printf("Finished transcription for %s in %v\n", chunkPath, time.Since(start))

	// Whisper CLI creates a .json file by stripping the original extension and adding .json
	// e.g., chunk_0001.wav -> chunk_0001.json
	filenameBase := filepath.Base(chunkPath)
	ext := filepath.Ext(filenameBase)
	jsonPath := filepath.Join(outputDir, filenameBase[:len(filenameBase)-len(ext)]+".json")
	defer os.Remove(jsonPath)

	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("reading transcription json: %w", err)
	}

	var result whisperOutput
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("parsing transcription output: %w\nRaw Output: %s", err, string(jsonData))
	}

	segments := make([]Segment, len(result.Segments))
	for i, seg := range result.Segments {
		segments[i] = Segment{
			Start:    offsetDuration + time.Duration(seg.Start*float64(time.Second)),
			End:      offsetDuration + time.Duration(seg.End*float64(time.Second)),
			Text:     seg.Text,
			Language: result.Language,
		}
	}

	return segments, nil
}
