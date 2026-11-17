package aiflags

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/jmylchreest/tinct/pkg/ai"
	"io"
	"strings"

	"github.com/jmylchreest/tinct/internal/plugin/input/shared/commonflags"
)

// ParseParams validates a single JSON object without losing integer precision.
func ParseParams() (map[string]any, error) {
	if strings.TrimSpace(Params) == "" {
		return map[string]any{}, nil
	}
	var params map[string]any
	decoder := json.NewDecoder(strings.NewReader(Params))
	decoder.UseNumber()
	if err := decoder.Decode(&params); err != nil {
		return nil, fmt.Errorf("--ai.params must be a JSON object: %w", err)
	}
	if params == nil {
		return nil, fmt.Errorf("--ai.params must be a JSON object, not null")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("--ai.params must contain exactly one JSON object")
	}
	return params, nil
}

// MergeParams recursively merges request-body overrides.
func MergeParams(dst, params map[string]any) map[string]any { return ai.MergeParams(dst, params) }

// MarshalRequest applies the CLI parameters to a provider request.
func MarshalRequest(request any) ([]byte, error) {
	params, err := ParseParams()
	if err != nil {
		return nil, err
	}
	return ai.MarshalRequest(request, params)
}

// CacheKey fingerprints generation and post-processing settings, excluding credentials.
// JSON map ordering is canonical, so whitespace/key order in Params does not matter.
func CacheKey(model string, providerSettings ...string) (string, error) {
	params, err := ParseParams()
	if err != nil {
		return "", err
	}
	raw, err := json.Marshal([]any{Prompt, model, params, NoExtendedPrompt,
		NoNegativePrompt, NegativePrompt, commonflags.AspectRatio,
		commonflags.TrimLetterbox, providerSettings})
	if err != nil {
		return "", fmt.Errorf("encode cache settings: %w", err)
	}
	hash := sha256.Sum256(raw)
	return hex.EncodeToString(hash[:])[:16], nil
}
