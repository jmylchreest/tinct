// Tinct - A modern colour palette generator
//
// Tinct extracts colour palettes from images and generates configuration
// files for your favorite applications.
//
// Copyright (c) 2025 John Mylchreest
// Licensed under the MIT License
package main

import (
	"os"

	"github.com/jmylchreest/tinct/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
