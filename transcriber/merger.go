package transcriber

import "os"

// mergeOverlappingSegments handles deduplication from overlapping chunks.
// Segments from overlapping regions are compared by timestamp;
// duplicates within the overlap window are dropped.
func mergeOverlappingSegments(segments []Segment) []Segment {
	if len(segments) == 0 {
		return segments
	}
	merged := []Segment{segments[0]}
	for i := 1; i < len(segments); i++ {
		prev := merged[len(merged)-1]
		curr := segments[i]
		// If curr starts before prev ends, it's from an overlap region.
		// Keep the one with the earlier start (from the first chunk),
		// as it has more left-context for accuracy.
		if curr.Start < prev.End {
			continue // skip duplicate from overlap
		}
		merged = append(merged, curr)
	}
	return merged
}
func cleanupTempFile(normalizedPath, originalPath string) {
	if normalizedPath != originalPath {
		os.Remove(normalizedPath)
	}
}
