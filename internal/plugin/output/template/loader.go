// Package template provides utilities for loading plugin templates with custom override support.
// This package re-exports everything from pkg/template; use that package for new code.
package template

import pkgtemplate "github.com/jmylchreest/tinct/pkg/template"

// Loader is an alias for pkg/template.Loader.
type Loader = pkgtemplate.Loader

// Logger is an alias for pkg/template.Logger.
type Logger = pkgtemplate.Logger

// Info is an alias for pkg/template.Info.
type Info = pkgtemplate.Info

// New creates a new template loader. Delegates to pkg/template.New.
var New = pkgtemplate.New

// NewFromFS creates a new template loader from an arbitrary fs.FS,
// used by the external-plugin path where the template tree is
// reconstructed from RPC bytes. Delegates to pkg/template.NewFromFS.
var NewFromFS = pkgtemplate.NewFromFS
