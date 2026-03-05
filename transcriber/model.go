package transcriber

import "time"

// Segment represents a single transcribed segment with timing info.
type Segment struct {
	Start    time.Duration `json:"start"`
	End      time.Duration `json:"end"`
	Text     string        `json:"text"`
	Language string        `json:"language,omitempty"`
}

// TranscriptionResult is the full output of a transcription job.
type TranscriptionResult struct {
	ID        string    `json:"id"`
	Filename  string    `json:"filename"`
	Segments  []Segment `json:"segments"`
	FullText  string    `json:"full_text"`
	Duration  float64   `json:"duration_seconds"`
	CreatedAt time.Time `json:"created_at"`
}

// TranscriptionRequest encapsulates input parameters.
type TranscriptionRequest struct {
	FilePath string `json:"file_path"`
	Language string `json:"language,omitempty"`
	Model    string `json:"model,omitempty"`
}
