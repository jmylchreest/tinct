package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	tinctplugin "github.com/jmylchreest/tinct/pkg/plugin"
)

var (
	// Version is the semantic version of the plugin.
	// Injected at build time via: -ldflags "-X main.Version=x.y.z"
	Version = "0.1.0"

	// Commit is the git commit hash of the build.
	// Injected at build time via: -ldflags "-X main.Commit=$(git rev-parse HEAD)"
	Commit = "unknown"

	// Date is the build date in RFC3339 format.
	// Injected at build time via: -ldflags "-X main.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
	Date = "unknown"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--plugin-info" {
		printPluginInfo()
		return
	}
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		printVersion()
		return
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Error,
		Output:     os.Stderr,
		JSONFormat: false,
	})

	outputPlugin := &Plugin{version: Version, commit: Commit, date: Date}

	pluginMap := map[string]plugin.Plugin{
		"output": &tinctplugin.OutputPluginRPC{Impl: outputPlugin},
	}

	tinctplugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: tinctplugin.Handshake,
		Plugins:         pluginMap,
		Logger:          logger,
	})
}

func printPluginInfo() {
	outputPlugin := &Plugin{version: Version, commit: Commit, date: Date}
	data, err := json.MarshalIndent(outputPlugin.GetMetadata(), "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling plugin info: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func printVersion() {
	fmt.Printf("tinct-plugin-awob %s\n", Version)
	fmt.Printf("Commit: %s\n", Commit)
	fmt.Printf("Built:  %s\n", Date)
}
