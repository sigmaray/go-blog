package main

import (
	"bytes"
	"html/template"
	"path/filepath"

	"go-blog/sanitize"
)

// loadHTMLTemplates parses all HTML templates and registers shared helpers.
// The include helper renders a named content template inside a layout shell.
func loadHTMLTemplates() *template.Template {
	var tmpl *template.Template

	funcs := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"subtract": func(a, b int) int {
			return a - b
		},
		"len": func(v interface{}) int {
			switch s := v.(type) {
			case []string:
				return len(s)
			default:
				return 0
			}
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(sanitize.HTML(s))
		},
		"include": func(name string, data interface{}) (template.HTML, error) {
			var buf bytes.Buffer
			if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
				return "", err
			}
			return template.HTML(buf.String()), nil
		},
	}

	tmpl = template.New("").Funcs(funcs)
	patterns := []string{
		filepath.Join("templates", "admin", "*.html"),
		filepath.Join("templates", "public", "*.html"),
	}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			panic(err)
		}
		if len(matches) == 0 {
			continue
		}
		tmpl = template.Must(tmpl.ParseFiles(matches...))
	}

	return tmpl
}
