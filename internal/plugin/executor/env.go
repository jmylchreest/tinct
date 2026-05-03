package executor

import (
	"os"
	"runtime"

	"github.com/jmylchreest/tinct/internal/version"
	"github.com/jmylchreest/tinct/pkg/plugin/paths"
)

// pluginEnv returns the environment slice to pass to an external plugin
// process. It inherits the parent environment and adds tinct-specific
// variables so plugins (in any language) can make portable decisions
// without re-implementing path resolution per-language.
//
// Variables exposed:
//
//   - TINCT_GOOS              runtime.GOOS (linux / darwin / windows)
//   - TINCT_GOARCH            runtime.GOARCH (amd64 / arm64 / ...)
//   - TINCT_VERSION           tinct's own version
//   - TINCT_HOME              resolved $HOME
//   - TINCT_XDG_CONFIG_HOME   resolved XDG config dir (per pkg/plugin/paths)
//   - TINCT_XDG_DATA_HOME     resolved XDG data dir
func pluginEnv() []string {
	env := os.Environ()

	home, _ := os.UserHomeDir() //nolint:errcheck // empty home is acceptable; plugins fall back to their own resolution

	env = append(env,
		"TINCT_GOOS="+runtime.GOOS,
		"TINCT_GOARCH="+runtime.GOARCH,
		"TINCT_VERSION="+version.Version,
		"TINCT_HOME="+home,
		"TINCT_XDG_CONFIG_HOME="+paths.XDGConfigDir(),
		"TINCT_XDG_DATA_HOME="+paths.XDGDataDir(),
	)
	return env
}
