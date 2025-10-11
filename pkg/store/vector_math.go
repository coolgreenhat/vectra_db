package store

import (
	"errors"
	"math"
	"strings"

	"github.com/agnivade/levenshtein"
)

// BM25 tuning parameters
const (
	k1 = 1.5 // term frequency scaling
	b  = 0.75 // document length normalization
)

// CosineSimilarity computes cosine similarity between two equal-length vectors.
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

// MetadataSimilarity: simple exact-match ratio for filters vs metadata
func MetadataSimilarity(filters, metadata map[string]string) float64 {
	if len(filters) == 0 {
		return 1.0
	}

	matchCount := 0
	for key, val := range filters {
		if metadata[key] == val {
			matchCount++
		}
	}
	return float64(matchCount) / float64(len(filters))
}

// tokenize a string (lowercase + split on whitespace).
func tokenize(text string) []string {
	parts := strings.Fields(strings.ToLower(text))
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.Trim(p, ".,!?\"'()[]{}:;")
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// termFrequency returns normalized TF (term count divided by doc length)
func termFrequency(tokens []string) map[string]float64 {
	tf := make(map[string]float64)
	if len(tokens) == 0 {
		return tf
	}
	for _, t := range tokens {
		tf[t]++
	}
	for t := range tf {
		tf[t] = tf[t] / float64(len(tokens))
	}
	return tf
}

// computeIDF computes IDF across the provided vectors (uses v.Text)
func computeIDF(vectors map[string]Vector) map[string]float64 {
	df := make(map[string]int)
	totalDocs := len(vectors)
	if totalDocs == 0 {
		return map[string]float64{}
	}

	for _, v := range vectors {
		seen := make(map[string]bool)
		for _, tok := range tokenize(v.Text) {
			if !seen[tok] {
				df[tok]++
				seen[tok] = true
			}
		}
	}

	idf := make(map[string]float64, len(df))
	for term, count := range df {
		idf[term] = math.Log(float64(totalDocs)/(1.0+float64(count))) // smoothing
	}
	return idf
}

// computeTFIDF multiplies TF by IDF for each term
func computeTFIDF(tf map[string]float64, idf map[string]float64) map[string]float64 {
	tfidf := make(map[string]float64, len(tf))
	for term, tfVal := range tf {
		tfidf[term] = tfVal * idf[term]
	}
	return tfidf
}

// cosineTFIDF computes cosine between two sparse TF-IDF maps
func cosineTFIDF(a, b map[string]float64) float64 {
	var dot, normA, normB float64
	for term, aVal := range a {
		bVal := b[term]
		dot += aVal * bVal
		normA += aVal * aVal
	}
	for _, bVal := range b {
		normB += bVal * bVal
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// FuzzySimilarity uses token-level Levenshtein best-match averaging (0..1)
func FuzzySimilarity(query, text string) float64 {
	qTokens := tokenize(query)
	tTokens := tokenize(text)
	if len(qTokens) == 0 || len(tTokens) == 0 {
		return 0
	}

	var totalScore float64
	for _, qt := range qTokens {
		best := 0.0
		for _, tt := range tTokens {
			dist := levenshtein.Distance(qt, tt)
			maxLen := len(qt)
			if len(tt) > maxLen {
				maxLen = len(tt)
			}
			if maxLen == 0 {
				continue
			}
			sim := 1.0 - (float64(dist) / float64(maxLen))
			if sim > best {
				best = sim
			}
		}
		totalScore += best
	}
	return totalScore / float64(len(qTokens))
}

// buildBM25Stats constructs per-doc term freq, avg doc length and idf map
func buildBM25Stats(docs []string) (docFreqs []map[string]int, avgDocLen float64, idf map[string]float64) {
	docFreqs = make([]map[string]int, len(docs))
	termDocCount := map[string]int{}
	totalLen := 0

	for i, doc := range docs {
		tokens := tokenize(doc)
		totalLen += len(tokens)

		freq := map[string]int{}
		seen := map[string]bool{}
		for _, t := range tokens {
			freq[t]++
			if !seen[t] {
				termDocCount[t]++
				seen[t] = true
			}
		}
		docFreqs[i] = freq
	}

	N := float64(len(docs))
	idf = make(map[string]float64, len(termDocCount))
	for term, df := range termDocCount {
		// BM25 IDF smoothing
		idf[term] = math.Log(1.0 + (N-float64(df)+0.5)/(float64(df)+0.5))
	}

	if N == 0 {
		avgDocLen = 0
	} else {
		avgDocLen = float64(totalLen) / N
	}
	return
}

// BM25Scores computes BM25 for each document, with optional fuzzy matching credit
func BM25Scores(query string, docs []string, fuzzyThreshold float64) []float64 {
	queryTerms := tokenize(query)
	docFreqs, avgDocLen, idf := buildBM25Stats(docs)

	scores := make([]float64, len(docs))
	for i, doc := range docs {
		freq := docFreqs[i]
		docTokens := tokenize(doc)
		docLen := float64(len(docTokens))
		score := 0.0

		for _, qTerm := range queryTerms {
			tf := float64(freq[qTerm])

			// If exact term not present, consider fuzzy matches and give partial credit
			if tf == 0 && fuzzyThreshold > 0 && len(docTokens) > 0 {
				for _, t := range docTokens {
					sim := 1.0 - float64(levenshtein.Distance(qTerm, t))/float64(maxInt(len(qTerm), len(t)))
					if sim >= fuzzyThreshold {
						tf += sim // partial credit proportional to similarity
					}
				}
			}

			if tf == 0 {
				continue
			}

			idfTerm := idf[qTerm] // if missing, idfTerm==0 (fine)
			norm := tf * (k1 + 1.0) / (tf + k1*(1.0-b+b*(docLen/avgDocLen)))
			score += idfTerm * norm
		}
		scores[i] = score
	}
	return scores
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
