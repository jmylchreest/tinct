// Package hooks provides a declarative pre/post-execute hook system for
// tinct output plugins. Plugins return a Spec describing routine
// behaviours (binary checks, dir auto-creation, reload commands,
// instructions messages); a shared runner evaluates the spec around the
// plugin's optional imperative PreExecute / PostExecute methods.
//
// Both internal plugins and external Go plugins import this package; the
// imperative hooks remain available for cases the spec can't express
// (e.g. KDE Plasma's variant-toggle workaround, GNOME Shell's User
// Themes extension state probe).
package hooks

import "context"

// Spec describes the routine pre/post-execute behaviour of a plugin.
// All fields are optional; an empty Spec is valid and means "no hooks".
type Spec struct {
	RequiredBinaries []string
	OptionalBinaries []string
	RequiredDirs     []string
	RequiredAny      []AnyOf
	AutoCreateDir    bool

	Reload   *ReloadSpec
	ReloadFn func(ctx context.Context) error

	MakeExecutable []string

	SupportsWallpaper bool
	Wallpaper         func(ctx context.Context, path string) error

	Instructions   string
	InstructionsFn func(ctx Context) string
}

// AnyOf is a satisfied-by-any group of detection candidates: the group
// passes when ANY of its Binaries or Dirs is present. Groups in
// Spec.RequiredAny are ANDed with each other (and with RequiredBinaries
// / RequiredDirs), so "all groups must be satisfied, each by at least
// one of its entries".
//
// This is the primitive for apps with more than one valid install
// shape — a config that may live in either of two places, or a tool
// detectable by config file OR by binary:
//
//	AnyOf{Dirs: []string{"~/.config/kdeglobals", "~/.config/plasmarc"}}
//	AnyOf{Binaries: []string{"rosec-prompt"}, Dirs: []string{"~/.config/rosec/config.toml"}}
//
// Binaries are resolved with appdetect (PATH, Flatpak, AppImage); Dirs
// accept files as well as directories and are ~-expanded. A single-entry
// group is legitimate when the point is to attach a better Reason than
// the runner's generated one.
//
// Reason, when non-empty, replaces the generated skip message. It may be
// multi-line — the runner indents continuation lines so install hints
// stay readable in a multi-plugin run.
type AnyOf struct {
	Binaries []string
	Dirs     []string
	Reason   string
}

// ReloadSpec describes a reload action as a verb plus arguments. Bare
// {Args: ["foo", "bar"]} (no Verb) is exec shorthand.
type ReloadSpec struct {
	Verb Verb
	Args []string
}

// Verb is the declarative reload verb. Add new verbs here when a routine
// reload pattern recurs across plugins.
type Verb string

const (
	VerbExec   Verb = "exec"
	VerbSignal Verb = "signal"
)

// Context is the execution context passed to InstructionsFn and to the
// runner. It mirrors the manager's notion of an execution context but
// stays inside the public SDK so plugins can implement Provider without
// importing internal packages.
//
// PluginName is the name of the plugin owning this execution; it is used
// to prefix verbose-mode messages (e.g. "kitty: Created directory: ...")
// so multi-plugin runs are easy to scan.
type Context struct {
	PluginName    string
	DryRun        bool
	Verbose       bool
	OutputDir     string
	WallpaperPath string
	WrittenFiles  []string
}

// Provider is implemented by plugins that opt into declarative hooks.
// The manager checks for this interface and runs the spec around the
// plugin's optional imperative PreExecute / PostExecute methods.
type Provider interface {
	Hooks() Spec
}
