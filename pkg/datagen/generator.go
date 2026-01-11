package datagen

import (
	"math/rand"
	"sync"
	"time"
)

const (
	Digits       = "0123456789"
	LatinLower   = "abcdefghijklmnopqrstuvwxyz"
	LatinUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	LatinLetters = LatinLower + LatinUpper
	Alphanumeric = LatinLetters + Digits
	SpecialChars = "`~!@#$%^&*()-_=+[]{}\\|;:'\",<.>/?"
)

const (
	defaultLength = 10
	emailDomain   = "@generated.com"
)

var (
	rng   *rand.Rand
	rngMu sync.Mutex
)

func init() {
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func String(length int, charsets ...string) string {
	if length <= 0 {
		length = defaultLength
	}

	charset := combineCharsets(charsets, Alphanumeric)
	if len(charset) == 0 {
		return ""
	}

	rngMu.Lock()
	defer rngMu.Unlock()

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[rng.Intn(len(charset))]
	}
	return string(result)
}

func Email(length int) string {
	if length <= 0 {
		length = defaultLength
	}
	return String(length, Alphanumeric) + emailDomain
}

func Password(length int, charsets ...string) string {
	if len(charsets) == 0 {
		charsets = []string{Digits, LatinUpper, LatinLower, SpecialChars}
	}

	var validCharsets []string
	for _, cs := range charsets {
		if len(cs) > 0 {
			validCharsets = append(validCharsets, cs)
		}
	}

	if len(validCharsets) == 0 {
		return ""
	}

	if length < len(validCharsets) {
		length = len(validCharsets)
	}

	rngMu.Lock()
	defer rngMu.Unlock()

	result := make([]byte, length)

	for i, charset := range validCharsets {
		result[i] = charset[rng.Intn(len(charset))]
	}

	allChars := ""
	for _, cs := range validCharsets {
		allChars += cs
	}

	for i := len(validCharsets); i < length; i++ {
		result[i] = allChars[rng.Intn(len(allChars))]
	}

	rng.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	return string(result)
}

func combineCharsets(charsets []string, defaultCharset string) string {
	if len(charsets) == 0 {
		return defaultCharset
	}
	result := ""
	for _, cs := range charsets {
		result += cs
	}
	if result == "" {
		return defaultCharset
	}
	return result
}
