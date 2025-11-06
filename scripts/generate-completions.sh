#!/bin/bash
# Generate shell completions for Tinct
# This script is run by GoReleaser before building releases

set -e

echo "Generating shell completions..."

# Create completions directory
mkdir -p completions

# Generate completions using Cobra's built-in completion commands
go run ./cmd/tinct completion bash > completions/tinct.bash
go run ./cmd/tinct completion zsh > completions/_tinct
go run ./cmd/tinct completion fish > completions/tinct.fish

echo "Generated completions:"
ls -lh completions/

echo "Shell completions generated successfully"
