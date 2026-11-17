package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// MergeParams recursively merges objects; arrays, scalars and null replace defaults.
// The destination is modified. The supplied params are not modified.
func MergeParams(dst, params map[string]any) map[string]any {
	for key, value := range params {
		if object, ok := value.(map[string]any); ok {
			base, isObject := dst[key].(map[string]any)
			if !isObject || base == nil {
				base = make(map[string]any)
			}
			dst[key] = MergeParams(base, object)
		} else {
			dst[key] = value
		}
	}
	return dst
}

// MarshalRequest applies explicit parameters to the final JSON request body.
func MarshalRequest(request any, params map[string]any) ([]byte, error) {
	raw, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("encode AI request: %w", err)
	}
	var body map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&body); err != nil || body == nil {
		return nil, fmt.Errorf("AI request must be a JSON object")
	}
	merged := MergeParams(body, params)
	if merged["stream"] == true {
		return nil, fmt.Errorf("stream=true is unsupported for image generation")
	}
	raw, err = json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("encode merged AI request: %w", err)
	}
	return raw, nil
}
