package transcriber

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hash-walker/volga-audio-transcriber/audio"
)

type Pipeline struct {
	engine     *WhisperEngine
	maxWorkers int
}

func NewPipeline(modelSize, device string, maxWorkers int) *Pipeline {
	if maxWorkers <= 0 {
		maxWorkers = 1
	}
	return &Pipeline{
		engine:     NewWhisperEngine(modelSize, device),
		maxWorkers: maxWorkers,
	}
}

// Run executes the full transcription pipeline.
func (p *Pipeline) Run(ctx context.Context, req TranscriptionRequest) (*TranscriptionResult, error) {
	// Step 1: Normalize audio format → 16kHz mono WAV

	normalizedPath, err := audio.NormalizeToWAV(req.FilePath)
	if err != nil {
		return nil, fmt.Errorf("normalization: %w", err)
	}

	defer cleanupTempFile(normalizedPath, req.FilePath)

	// Step 2: Get total duration for the result
	totalDuration, err := audio.GetDuration(normalizedPath)
	if err != nil {
		return nil, fmt.Errorf("duration check: %w", err)
	}

	// Step 3: Split into chunks (no-op for short files)

	chunkDir, _ := os.MkdirTemp("", "transcribe-chunks-*")
	defer os.RemoveAll(chunkDir)
	chunks, err := audio.SplitAudio(normalizedPath, chunkDir)

	if err != nil {
		return nil, fmt.Errorf("chunking: %w", err)
	}

	// Step 4: Transcribe chunks (concurrently if multiple workers)
	allSegments, err := p.transcribeChunks(ctx, chunks)

	if err != nil {
		return nil, fmt.Errorf("transcription: %w", err)
	}

	// Step 5: Merge overlapping segments & deduplicate
	mergedSegments := mergeOverlappingSegments(allSegments)
	
	// Step 6: Build final result
	var fullText strings.Builder
	for i, seg := range mergedSegments {
		if i > 0 {
			fullText.WriteString(" ")
		}
		fullText.WriteString(strings.TrimSpace(seg.Text))
	}
	return &TranscriptionResult{
		ID:        uuid.New().String(),
		Filename:  req.FilePath,
		Segments:  mergedSegments,
		FullText:  fullText.String(),
		Duration:  totalDuration.Seconds(),
		CreatedAt: time.Now(),
	}, nil
}

func (p *Pipeline) transcribeChunks(ctx context.Context, chunks []audio.AudioChunk) ([]Segment, error) {
	type chunkResult struct {
		index    int
		segments []Segment
		err      error
	}

	results := make(chan chunkResult, len(chunks))
	sem := make(chan struct{}, p.maxWorkers)

	var wg sync.WaitGroup

	for _, chunk := range chunks {
		wg.Add(1)

		go func(c audio.AudioChunk) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				results <- chunkResult{index: c.Index, err: ctx.Err()}
				return
			case sem <- struct{}{}:
				defer func() { <-sem }()
			}

			segments, err := p.engine.TranscribeChunk(c.Path, c.Start)
			results <- chunkResult{index: c.Index, segments: segments, err: err}

		}(chunk)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results in order
	ordered := make([][]Segment, len(chunks))
	for res := range results {
		if res.err != nil {
			return nil, fmt.Errorf("chunk %d: %w", res.index, res.err)
		}

		ordered[res.index] = res.segments
	}
	// Flatten
	var all []Segment
	for _, segs := range ordered {
		all = append(all, segs...)
	}
	return all, nil

}
