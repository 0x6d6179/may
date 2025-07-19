package fuzzy

import "strings"

// Score returns a score between 0.0 and 1.0 for how well query matches target.
// It returns 1.0 for an exact match, 0.0 for completely different strings.
func Score(query, target string) float64 {
	query = strings.ToLower(query)
	target = strings.ToLower(target)

	if query == target {
		return 1.0
	}
	if len(query) == 0 {
		return 0.0
	}
	if len(target) == 0 {
		return 0.0
	}

	// Fast path for prefix match
	if strings.HasPrefix(target, query) {
		// e.g. target "dashboard", query "dash" -> 0.9 to 0.99 depending on length
		// Give prefix matches a very high base score
		return 0.9 + (0.1 * float64(len(query)) / float64(len(target)))
	}

	// Fast path for contains
	if strings.Contains(target, query) {
		return 0.8 + (0.1 * float64(len(query)) / float64(len(target)))
	}

	// Calculate Levenshtein distance
	d := levenshtein(query, target)
	maxLen := len(query)
	if len(target) > maxLen {
		maxLen = len(target)
	}

	similarity := 1.0 - float64(d)/float64(maxLen)
	
	// If it contains the characters in order (subsequence), boost it
	if isSubsequence(query, target) {
		similarity = similarity*0.5 + 0.5 // Boost score if it's a subsequence
	}

	return similarity
}

func isSubsequence(query, target string) bool {
	i, j := 0, 0
	for i < len(query) && j < len(target) {
		if query[i] == target[j] {
			i++
		}
		j++
	}
	return i == len(query)
}

func levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	v0 := make([]int, len(b)+1)
	v1 := make([]int, len(b)+1)

	for i := 0; i <= len(b); i++ {
		v0[i] = i
	}

	for i := 0; i < len(a); i++ {
		v1[0] = i + 1
		for j := 0; j < len(b); j++ {
			cost := 1
			if a[i] == b[j] {
				cost = 0
			}
			v1[j+1] = min(v1[j]+1, v0[j+1]+1, v0[j]+cost)
		}
		for j := 0; j <= len(b); j++ {
			v0[j] = v1[j]
		}
	}

	return v1[len(b)]
}

func min(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}
