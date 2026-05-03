package executor

import (
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestPluginEnv_IncludesTinctVars(t *testing.T) {
	env := pluginEnv()

	required := []string{
		"TINCT_GOOS=" + runtime.GOOS,
		"TINCT_GOARCH=" + runtime.GOARCH,
	}
	for _, want := range required {
		if !slices.Contains(env, want) {
			t.Errorf("env missing required var %q", want)
		}
	}

	// Variables that should be set with non-empty values
	prefixes := []string{
		"TINCT_VERSION=",
		"TINCT_HOME=",
		"TINCT_XDG_CONFIG_HOME=",
		"TINCT_XDG_DATA_HOME=",
	}
	for _, prefix := range prefixes {
		found := false
		for _, e := range env {
			if strings.HasPrefix(e, prefix) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("env missing variable with prefix %q", prefix)
		}
	}
}

func TestPluginEnv_PreservesParentEnv(t *testing.T) {
	t.Setenv("TINCT_TEST_INHERITED_VAR", "hello")
	env := pluginEnv()
	if !slices.Contains(env, "TINCT_TEST_INHERITED_VAR=hello") {
		t.Error("pluginEnv should inherit parent environment variables")
	}
}
