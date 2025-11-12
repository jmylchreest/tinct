package repomanager

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Verifier handles URL verification and checksum calculation.
type Verifier struct {
	client  *http.Client
	timeout time.Duration
}

// NewVerifier creates a new URL verifier.
func NewVerifier() *Verifier {
	return &Verifier{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		timeout: 30 * time.Second,
	}
}

// VerifyURL checks if a URL is accessible.
func (v *Verifier) VerifyURL(url string) (available bool, reason string) {
	resp, err := v.client.Head(url)
	if err != nil {
		return false, fmt.Sprintf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, ""
	}

	return false, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status)
}

// CalculateChecksum downloads a file and calculates its SHA256 checksum.
func CalculateChecksum(url string) (checksum string, size int64, err error) {
	resp, err := http.Get(url) // #nosec G107 - URL is from user input, validated upstream
	if err != nil {
		return "", 0, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Calculate SHA256 while reading
	hash := sha256.New()
	size, err = io.Copy(hash, resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("read failed: %w", err)
	}

	checksum = fmt.Sprintf("%x", hash.Sum(nil))
	return checksum, size, nil
}

// VerifyChecksum downloads a file and verifies its checksum.
func VerifyChecksum(url, expectedChecksum string) (bool, error) {
	actualChecksum, _, err := CalculateChecksum(url)
	if err != nil {
		return false, err
	}

	return actualChecksum == expectedChecksum, nil
}
