package store

import (
	"errors"
	"math"
)

func CosineSimilarity(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, errors.New("vectors must have the same length")
	}

	var dot, magA, magB float64
	for i := range a {
		dot += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}

	if magA == 0 || magB == 0 {
		return 0, errors.New("zero-length vector")
	}

	return dot / (math.Sqrt(magA) * math.Sqrt(magB)), nil
}
