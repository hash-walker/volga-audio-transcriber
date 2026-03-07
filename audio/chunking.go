package audio

import (
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	// ChunkDuration is the max duration per chunk.
	// 30s is Whisper's native window; keeps memory bounded

	ChunkDuration = 30 * time.Second

	// OverlapDuration prevents word-cutting at chunk boundaries
	// 2s overlap lets the model capture words that straddle boundaries

	OverlapDuration = 2 * time.Second
)

type AudioChunk struct {
	Path  string
	Start time.Duration
	End   time.Duration
	Index int
}

func SplitAudio(inputPath, chunkDir string) ([]AudioChunk, error) {
	totalDuration, err := GetDuration(inputPath)

	if err != nil {
		return nil, fmt.Errorf("failed to get duration: %w", err)
	}

	// Short files don't need chunking
	if totalDuration <= ChunkDuration {
		return []AudioChunk{{
			Path:  inputPath,
			Start: 0,
			End:   totalDuration,
			Index: 0,
		}}, nil
	}

	var chunks []AudioChunk
	step := ChunkDuration - OverlapDuration
	numChunks := int(math.Ceil(float64(totalDuration) / float64(step)))

	for i := 0; i < numChunks; i++ {
		start := time.Duration(i) * step
		end := start + ChunkDuration

		if end > totalDuration {
			end = totalDuration
		}

		chunkPath := fmt.Sprintf("%s/chunk_%04d.wav", chunkDir, i)

		cmd := exec.Command("ffmpeg",
			"-i", inputPath,
			"-ss", formatDuration(start),
			"-t", formatDuration(end-start),
			"-y",
			chunkPath,
		)

		if output, chunkErr := cmd.CombinedOutput(); chunkErr != nil {
			return nil, fmt.Errorf("chunk %d failed: %w\n%s", i, chunkErr, string(output))
		}

		chunks = append(chunks, AudioChunk{
			Path:  chunkPath,
			Start: start,
			End:   end,
			Index: i,
		})

	}

	return chunks, nil
}

// GetDuration returns duration of the audio using ffprobe

func GetDuration(path string) (time.Duration, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)

	output, err := cmd.Output()

	if err != nil {
		return 0, err
	}

	seconds, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)

	if err != nil {
		return 0, err
	}

	return time.Duration(seconds * float64(time.Second)), nil
}

func formatDuration(d time.Duration) string {
	totalSeconds := d.Seconds()
	hours := int(totalSeconds) / 3600
	minutes := (int(totalSeconds) % 3600) / 60
	seconds := totalSeconds - float64(hours*3600) - float64(minutes*60)
	return fmt.Sprintf("%02d:%02d:%06.3f", hours, minutes, seconds)
}
