// Package aiflags provides shared flags for AI image generation plugins.
// These flags are common across all AI-based input plugins like googlegenai and openrouter.
package aiflags

import (
	"sync"

	"github.com/spf13/cobra"
)

// Shared AI flag variables - these are shared across all AI plugins.
var (
	// Prompt is the text description for image generation.
	Prompt string

	// Model is the AI model to use for generation.
	Model string

	// ListModels indicates whether to list available models and exit.
	ListModels bool

	// NoExtendedPrompt disables automatic wallpaper prompt enhancements.
	NoExtendedPrompt bool

	// NoNegativePrompt disables the default negative prompt.
	NoNegativePrompt bool

	// NegativePrompt is a custom negative prompt to discourage certain elements.
	NegativePrompt string

	mu sync.Mutex
)

// RegisterFlags registers the shared AI flags on the given command.
// This should be called by each AI plugin - it will only register once.
// Returns true if this call performed the registration, false if already registered.
func RegisterFlags(cmd *cobra.Command) bool {
	mu.Lock()
	defer mu.Unlock()

	// Check if already registered on this command
	if cmd.Flags().Lookup("ai.prompt") != nil {
		return false
	}

	cmd.Flags().StringVar(&Prompt, "ai.prompt", "", "Text description for AI image generation (required for AI plugins)")
	cmd.Flags().StringVar(&Model, "ai.model", "auto", "AI model to use for image generation")
	cmd.Flags().BoolVar(&ListModels, "ai.list-models", false, "List available AI models and exit")
	cmd.Flags().BoolVar(&NoExtendedPrompt, "ai.no-extended-prompt", false, "Disable automatic wallpaper prompt enhancements")
	cmd.Flags().BoolVar(&NoNegativePrompt, "ai.no-negative-prompt", false, "Disable default negative prompt")
	cmd.Flags().StringVar(&NegativePrompt, "ai.negative-prompt", "", "Custom negative prompt to discourage certain elements")

	return true
}

// Reset resets all flag values to defaults. Useful for testing.
func Reset() {
	mu.Lock()
	defer mu.Unlock()

	Prompt = ""
	Model = "auto"
	ListModels = false
	NoExtendedPrompt = false
	NoNegativePrompt = false
	NegativePrompt = ""
}
