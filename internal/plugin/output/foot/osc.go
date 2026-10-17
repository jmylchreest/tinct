package foot

import (
	"fmt"
	"regexp"
	"strings"
)

var terminalPath = regexp.MustCompile(`^/dev/pts/\d+$`)

var hexColour = regexp.MustCompile(`^[0-9a-fA-F]{6}$`)

// paletteOSC uses the rendered palette, including custom template overrides.
// Only validated RGB values enter terminal control sequences.
func paletteOSC(data []byte) ([]byte, error) {
	values := map[string]string{}
	section := ""
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			section = line
			continue
		}
		if section != "[colors]" {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if ok {
			values[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	var out strings.Builder
	appendColour := func(code, key string) error {
		value := values[key]
		if !hexColour.MatchString(value) {
			return fmt.Errorf("foot live palette: %s must be six hex digits", key)
		}
		fmt.Fprintf(&out, "\x1b]%s;#%s\x1b\\", code, value)
		return nil
	}
	for i := range 16 {
		key := fmt.Sprintf("regular%d", i)
		if i >= 8 {
			key = fmt.Sprintf("bright%d", i-8)
		}
		if err := appendColour(fmt.Sprintf("4;%d", i), key); err != nil {
			return nil, err
		}
	}
	for _, entry := range []struct{ code, key string }{
		{"10", "foreground"}, {"11", "background"}, {"12", "foreground"},
		{"17", "selection-background"}, {"19", "selection-foreground"},
	} {
		if err := appendColour(entry.code, entry.key); err != nil {
			return nil, err
		}
	}
	return []byte(out.String()), nil
}
