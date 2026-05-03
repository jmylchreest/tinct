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
	AutoCreateDir    bool

	Reload   *ReloadSpec
	ReloadFn func(ctx context.Context) error

	MakeExecutable []string

	SupportsWallpaper bool
	Wallpaper         func(ctx context.Context, path string) error

	Instructions   string
	InstructionsFn func(ctx Context) string
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
