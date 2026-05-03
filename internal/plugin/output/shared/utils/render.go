package utils

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"strings"
	"text/template"

	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
)

// RenderTemplate loads pluginName's templateFile via tmplloader (which honours
// user template overrides on disk), parses it with TemplateFuncs(), and
// executes it against data. It returns the rendered bytes and a flag
// indicating whether the loaded source came from a user customisation; the
// flag is intended for verbose logging by callers.
//
// templateFile is the basename inside embedFS (e.g. "tinct.colors.tmpl"); the
// text/template name is the same with any .tmpl suffix stripped. verbose
// toggles progress logging inside tmplloader.
func RenderTemplate(pluginName, templateFile string, embedFS embed.FS, data any, verbose bool) (rendered []byte, fromCustom bool, err error) {
	loader := tmplloader.New(pluginName, embedFS)
	if verbose {
		loader.WithVerbose(true, NewVerboseLogger(os.Stderr))
	}

	content, fromCustom, err := loader.Load(templateFile)
	if err != nil {
		return nil, false, fmt.Errorf("failed to load template %s: %w", templateFile, err)
	}

	name := strings.TrimSuffix(templateFile, ".tmpl")
	tmpl, err := template.New(name).Funcs(TemplateFuncs()).Parse(string(content))
	if err != nil {
		return nil, fromCustom, fmt.Errorf("failed to parse template %s: %w", templateFile, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fromCustom, fmt.Errorf("failed to execute template %s: %w", templateFile, err)
	}

	return buf.Bytes(), fromCustom, nil
}
