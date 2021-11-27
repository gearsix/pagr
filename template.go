package main

import (
	"path/filepath"
	"io/fs"
	"strings"
	"notabug.org/gearsix/suti"
)

// DefaultTemplateName provides the default name for the template used
// when one isn't specified in a `Page.Meta`.
const DefaultTemplateName = "root"

// LoadTemplateDir loads all files in `dir` that are not directories as a `suti.Template`
// by calling `suti.LoadTemplateFile`. Partials for each template will be parsed from all
// files in a directory matching the base filename of the template (not including
// extension) if it exists.
func LoadTemplateDir(dir string) (templates []suti.Template, err error) {
	templatePaths := make(map[string][]string) // map[rootPath][]partialPaths...

	err = filepath.Walk(dir, func(path string, info fs.FileInfo, e error) error {
		lang := strings.TrimPrefix(filepath.Ext(path), ".")
		if e != nil || info.IsDir() || strings.Contains(path, ".ignore") || suti.IsSupportedTemplateLang(lang) == -1 {
			return e
		}

		templatePaths[path] = make([]string, 0)
		return e
	})

	err = filepath.Walk(dir, func(path string, info fs.FileInfo, e error) error {
		lang := strings.TrimPrefix(filepath.Ext(path), ".")
		if e != nil || info.IsDir() || ignoreFile(path) || suti.IsSupportedTemplateLang(lang) == -1 {
			return e
		}

		for t, _ := range templatePaths {
			if strings.Contains(path, filepath.Dir(t)) &&
				filepath.Ext(t) == filepath.Ext(path) {
					templatePaths[t] = append(templatePaths[t], path)
				}
		}
		return e
	})


	if err == nil {
		var t suti.Template
		for rootPath, partialPaths := range templatePaths {
			t, err = suti.LoadTemplateFilepath(rootPath, partialPaths...)
			if err != nil {
				break
			}
			templates = append(templates, t)
		}
	}

	return
}
