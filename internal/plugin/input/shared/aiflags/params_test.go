package aiflags

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/plugin/input/shared/commonflags"
)

func TestParseParams(t *testing.T) {
	defer Reset()
	for _, raw := range []string{"null", "[]", "42", `{"x":1} {}`, `{"x":`, `{"x":1} garbage`} {
		Params = raw
		if _, err := ParseParams(); err == nil {
			t.Errorf("accepted %q", raw)
		}
	}
	for _, raw := range []string{"", "  ", "{}", `{"seed":9007199254740993,"enabled":false,"items":[1,null,"x"]}`} {
		Params = raw
		if _, err := ParseParams(); err != nil {
			t.Errorf("rejected %q: %v", raw, err)
		}
	}
}

func TestMarshalRequestMerge(t *testing.T) {
	defer Reset()
	Params = `{"config":{"quality":"high","new":false},"list":[3],"seed":9007199254740993,"drop":null}`
	raw, err := MarshalRequest(map[string]any{"config": map[string]any{"quality": "low", "size": "2K"}, "list": []int{1, 2}, "drop": true})
	if err != nil {
		t.Fatal(err)
	}
	expected := `{"config":{"new":false,"quality":"high","size":"2K"},"drop":null,"list":[3],"seed":9007199254740993}`
	if string(raw) != expected {
		t.Fatalf("got %s, want %s", raw, expected)
	}
	if _, err := MarshalRequest([]string{"bad"}); err == nil {
		t.Fatal("accepted non-object request")
	}
	if _, err := MarshalRequest(make(chan int)); err == nil {
		t.Fatal("accepted unencodable request")
	}
}

func TestCacheKeySettings(t *testing.T) {
	Reset()
	commonflags.Reset()
	defer Reset()
	defer commonflags.Reset()
	Prompt = "wallpaper"
	Params = `{"quality":"high","seed":12}`
	first, err := CacheKey("model", "endpoint")
	if err != nil {
		t.Fatal(err)
	}
	Params = `{ "seed":12, "quality":"high" }`
	same, _ := CacheKey("model", "endpoint")
	if first != same {
		t.Fatal("JSON ordering changed cache key")
	}
	for _, change := range []func(){func() { Params = `{"quality":"low"}` }, func() { NoNegativePrompt = true }, func() { commonflags.AspectRatio = "1:1" }, func() { commonflags.TrimLetterbox = !commonflags.TrimLetterbox }} {
		before, _ := CacheKey("model", "endpoint")
		change()
		after, _ := CacheKey("model", "endpoint")
		if before == after {
			t.Fatal("settings change did not change key")
		}
	}
	a, _ := CacheKey("model", "a")
	b, _ := CacheKey("model", "b")
	if a == b {
		t.Fatal("endpoint missing from cache key")
	}
	Params = "[]"
	if _, err := CacheKey("model"); err == nil {
		t.Fatal("invalid params accepted")
	}
}

func TestMergeDoesNotMutateParams(t *testing.T) {
	p := map[string]any{"config": map[string]any{"quality": "high"}}
	dst := MergeParams(map[string]any{}, p)
	dst["config"].(map[string]any)["quality"] = "low"
	raw, _ := json.Marshal(p)
	if !strings.Contains(string(raw), "high") {
		t.Fatal("input mutated")
	}
}
