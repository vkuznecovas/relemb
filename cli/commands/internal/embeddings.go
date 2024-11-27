package internal

import (
	"fmt"
	"math"
)

func CosineSimilarity(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vectors must be of the same length")
	}

	dotProduct := 0.0
	magA := 0.0
	magB := 0.0

	for i := range a {
		dotProduct += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}

	if magA == 0 || magB == 0 {
		return 0, fmt.Errorf("magnitude of one or both vectors is zero")
	}

	magnitude := math.Sqrt(magA) * math.Sqrt(magB)
	return dotProduct / magnitude, nil
}
