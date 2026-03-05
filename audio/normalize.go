package audio

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

var SupportedFormats = map[string]bool{
	".wav":  true,
	".mp3":  true,
	".flac": true,
	".ogg":  true,
	".m4a":  true,
	".webm": true,
}

func NormalizeToWAV(inputPath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(inputPath))

	if !SupportedFormats[ext] {
		return "", fmt.Errorf("unsupported audio format: %s", ext)
	}

	outputPath := strings.TrimSuffix(inputPath, ext) + "_normalised.wav"

	if ext == ".wav" && isCorrectFormat(inputPath) {
		return outputPath, nil
	}

	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-vn",
		"-ar", "16000",
		"-ac", "1",
		"-sample_fmt", "s16",
		"-y",
		outputPath,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("ffmpeg conversion failed: %w\nOutput: %s", err, string(output))
	}

	return outputPath, nil

}

func isCorrectFormat(path string) bool {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=sample_rate,channels,sample_fmt",
		"-of", "csv=p=0",
		"-sample_fmt", "s16",
		path,
	)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Output looks like: "16000,1,s16" or "16000,1,s16p"
	// We trim space and check the string.
	res := strings.TrimSpace(string(output))

	// Check for 16kHz, 1 Channel, and s16 (or s16p planar) format
	return strings.HasPrefix(res, "16000,1,s16")
}
