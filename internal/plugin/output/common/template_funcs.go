// Package common provides shared utilities for output plugins.
package common

import (
	"text/template"

	pkgtemplate "github.com/jmylchreest/tinct/pkg/template"
)

// TemplateFuncs returns standard template functions for all output plugins.
// This delegates to the public pkg/template.TemplateFuncs() to avoid duplication.
// Both internal and external plugins use the same template functions for consistency.
func TemplateFuncs() template.FuncMap {
	// Simply re-export the public template functions.
	// Since internal/colour.ThemeData embeds pkg/colour.ThemeData,
	// all the public functions work seamlessly with internal types.
	return pkgtemplate.TemplateFuncs()
}
