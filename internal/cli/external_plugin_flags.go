// External plugins don't have their flags compiled into the tinct
// binary, so cobra can't register --<plugin>.<arg> entries for them
// statically the way it does for built-ins. This file walks the
// install manifest at command init, asks each external plugin for
// its flag list, and registers each as a real cobra flag; the
// collector merges parsed values back into generatePluginArgs so the
// existing executor path keeps working unchanged.
//
// The --plugin-args 'paletty={"palette":"…"}' JSON form remains as
// an escape hatch (and as the fallback for json-stdio plugins, which
// don't expose flag-help over the wire today).
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/plugin/executor"
	"github.com/jmylchreest/tinct/internal/plugin/input"
)

// externalPluginFlagMapping records which cobra flag name corresponds
// to which (plugin, arg) pair so the post-parse collector can route
// values into the right --plugin-args JSON entry.
type externalPluginFlagMapping struct {
	plugin  string
	arg     string
	argType string // "string" | "bool" | "int" — from FlagHelp.Type
}

// externalPluginFlags maps the cobra flag name (e.g. "paletty.palette")
// to its mapping. Populated by registerExternalPluginFlags, read by
// collectExternalPluginFlags.
var externalPluginFlags = map[string]externalPluginFlagMapping{}

// registerExternalPluginFlags walks the install manifest, queries
// each external plugin's flag help, and registers each as
// --<plugin>.<arg> on cmd. Skipped for commands that don't use
// plugins (avoids spawning --plugin-info during `tinct version`
// etc.). json-stdio plugins return no flags via the executor today
// and fall through to --plugin-args.
func registerExternalPluginFlags(cmd *cobra.Command) {
	if !commandUsesPlugins() {
		return
	}
	lock, _, err := loadPluginManifest()
	if err != nil || lock == nil {
		return
	}
	for _, meta := range lock.LivePlugins() {
		flags := getPluginFlagHelp(meta)
		if len(flags) == 0 {
			continue
		}
		for _, f := range flags {
			if f.Name == "" {
				continue
			}
			cobraName := fmt.Sprintf("%s.%s", meta.Name, f.Name)
			if cmd.Flags().Lookup(cobraName) != nil {
				continue
			}
			description := fmt.Sprintf("(%s) %s", meta.Name, f.Description)
			switch strings.ToLower(f.Type) {
			case "bool":
				cmd.Flags().Bool(cobraName, parseBoolOr(f.Default, false), description)
			case "int":
				cmd.Flags().Int(cobraName, parseIntOr(f.Default, 0), description)
			default:
				cmd.Flags().String(cobraName, f.Default, description)
			}
			externalPluginFlags[cobraName] = externalPluginFlagMapping{
				plugin:  meta.Name,
				arg:     f.Name,
				argType: strings.ToLower(f.Type),
			}
		}
	}
}

// collectExternalPluginFlags merges any dynamic --<plugin>.<arg>
// values cobra parsed into pluginArgs (the same map cobra populates
// from --plugin-args). Dynamic flags and explicit --plugin-args
// entries for the same plugin are unioned, dynamic-flag values win
// on key collision.
func collectExternalPluginFlags(cmd *cobra.Command, pluginArgs map[string]string) map[string]string {
	if pluginArgs == nil {
		pluginArgs = map[string]string{}
	}
	for cobraName, mapping := range externalPluginFlags {
		if !cmd.Flags().Changed(cobraName) {
			continue
		}
		var value any
		// The flag was registered above with the matching type, so the
		// type-check inside cobra's GetBool/GetInt/GetString cannot fail.
		switch mapping.argType {
		case "bool":
			value, _ = cmd.Flags().GetBool(cobraName) //nolint:errcheck
		case "int":
			value, _ = cmd.Flags().GetInt(cobraName) //nolint:errcheck
		default:
			value, _ = cmd.Flags().GetString(cobraName) //nolint:errcheck
		}
		merged, err := mergeArgIntoJSON(pluginArgs[mapping.plugin], mapping.arg, value)
		if err != nil {
			// JSON corruption from the user's explicit --plugin-args is
			// surfaced loudly. The dynamic-flag side is always
			// well-formed, so a failure here points at the explicit
			// input.
			fmt.Fprintf(os.Stderr, "Warning: could not merge --%s into --plugin-args %s: %v\n",
				cobraName, mapping.plugin, err)
			continue
		}
		pluginArgs[mapping.plugin] = merged
	}
	return pluginArgs
}

// mergeArgIntoJSON parses existing as a JSON object (empty string OK),
// sets [arg] = value, and re-serialises. Used to combine dynamic
// --<plugin>.<arg> flag values with whatever the user passed via
// --plugin-args for the same plugin.
func mergeArgIntoJSON(existing, arg string, value any) (string, error) {
	args := map[string]any{}
	if existing != "" {
		if err := json.Unmarshal([]byte(existing), &args); err != nil {
			return "", fmt.Errorf("existing plugin-args entry is not valid JSON: %w", err)
		}
	}
	args[arg] = value
	out, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("re-serialising plugin-args: %w", err)
	}
	return string(out), nil
}

// commandUsesPlugins sniffs os.Args to skip plugin discovery for
// unrelated commands. First non-flag positional arg decides.
func commandUsesPlugins() bool {
	for _, a := range os.Args[1:] {
		if strings.HasPrefix(a, "-") {
			continue
		}
		switch a {
		case "generate", "extract":
			return true
		case "help":
			// `tinct help generate` should also register so the dynamic
			// flags render in the help text; keep scanning.
			continue
		default:
			return false
		}
	}
	return false
}

// getPluginFlagHelp returns an external plugin's flag list. Prefers
// the cached value on meta.Flags (populated at install time) and
// falls back to spawning the binary when the cache is empty (e.g.
// plugins installed before the cache landed, or after a binary swap).
func getPluginFlagHelp(meta *ExternalPluginMeta) []input.FlagHelp {
	if meta == nil {
		return nil
	}
	if len(meta.Flags) > 0 {
		return meta.Flags
	}
	return fetchPluginFlags(meta.Path)
}

// PopulateFlagsCache fetches the plugin's flag help and stores it on
// meta.Flags. Install/update sites call this before saving the
// manifest so the runtime path doesn't have to spawn on every
// invocation.
func PopulateFlagsCache(meta *ExternalPluginMeta) {
	if meta == nil || meta.Path == "" {
		return
	}
	if flags := fetchPluginFlags(meta.Path); len(flags) > 0 {
		meta.Flags = flags
	}
}

func fetchPluginFlags(path string) []input.FlagHelp {
	if path == "" {
		return nil
	}
	exec, err := executor.NewWithVerbose(path, false)
	if err != nil {
		return nil
	}
	defer exec.Close()
	flags, err := exec.GetFlagHelp(context.Background())
	if err != nil {
		return nil
	}
	return flags
}

func parseBoolOr(s string, fallback bool) bool {
	if s == "" {
		return fallback
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return fallback
	}
	return v
}

func parseIntOr(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}
