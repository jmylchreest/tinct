// Package cli provides the command-line interface for Tinct.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/plugin/manager"
	"github.com/jmylchreest/tinct/internal/version"
)

var (
	// Global theme flag.
	globalTheme string

	// Global desaturation flag for alternate themes.
	globalNoDesaturation bool

	// Shared plugin manager instance used by all commands.
	sharedPluginManager *manager.Manager

	// RootCmd represents the base command when called without any subcommands.
	RootCmd = &cobra.Command{
		Use:   "tinct",
		Short: "A modern color palette generator",
		Long: `Tinct is a modern, extensible CLI tool that extracts color palettes from images
and generates configuration files for your favorite applications.

Extract vibrant color schemes from wallpapers and apply them system-wide to
terminal emulators, window managers, application launchers, and more.`,
		Version:      version.Short(),
		SilenceUsage: true,
	}
)

// NewRootCmd adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func NewRootCmd() *cobra.Command {
	// Initialise shared plugin manager.
	sharedPluginManager = manager.NewBuilder().Build()

	// Register plugin flags with all commands that need them.
	registerPluginFlags()

	// Global flags.
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")
	RootCmd.PersistentFlags().BoolP("quiet", "q", false, "suppress non-error output")
	RootCmd.PersistentFlags().StringVarP(&globalTheme, "theme", "t", "auto", "theme type (auto, dark, light)")
	RootCmd.PersistentFlags().BoolVar(&globalNoDesaturation, "no-desaturation", false, "disable desaturation when generating alternate themes (keeps original color saturation)")

	// Set version template.
	RootCmd.SetVersionTemplate(version.String() + "\n")

	// Add subcommands.
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(extractCmd)
	RootCmd.AddCommand(generateCmd)
	RootCmd.AddCommand(pluginsCmd)
	RootCmd.AddCommand(filesCmd)

	return RootCmd
}

func init() {
	// Tear down cached external-plugin executors at the end of every
	// CLI invocation; otherwise plugin subprocesses leak across runs.
	cobra.OnFinalize(func() {
		if sharedPluginManager != nil {
			sharedPluginManager.CloseAllExternalPlugins()
		}
	})
}

// registerPluginFlags registers plugin-specific flags with commands that use them.
func registerPluginFlags() {
	// Register input plugin flags.
	for _, plugin := range sharedPluginManager.AllInputPlugins() {
		plugin.RegisterFlags(extractCmd)
		plugin.RegisterFlags(generateCmd)
	}

	// Register output plugin flags.
	for _, plugin := range sharedPluginManager.AllOutputPlugins() {
		plugin.RegisterFlags(generateCmd)
	}
}

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print detailed version information including build date, commit hash, and Go version.`,
	Run: func(_ *cobra.Command, _ []string) {
		info := version.GetInfo()

		// Print version information in a structured format.
		fmt.Printf("Version:    %s\n", info.Version)
		fmt.Printf("Commit:     %s\n", info.Commit)
		fmt.Printf("Build Date: %s\n", info.Date)
		fmt.Printf("Go Version: %s\n", info.GoVersion)
		fmt.Printf("Platform:   %s\n", info.Platform)
	},
}
