package main

import (
	"path/filepath"
	"io/fs"
	"strings"
	"os"
	"notabug.org/gearsix/suti"
)

// DefaultTemplate provides the default name for the template used
// when one isn't specified in a `Page.Meta`.
const DefaultTemplate = "root"

// loadPaths calls `filepath.Walk` on dir and loads all
// non-directory filepaths in `dir`
func loadPaths(dir string) ([]string, error) {
	var r []string
	err := filepath.Walk(dir,
		func(fpath string, info fs.FileInfo, e error) error {
			if e != nil {
				return e
			}
			if !info.IsDir() {
				r = append(r, fpath)
			}
			return e
	})
	return r, err
}

// LoadTemplateDir loads all files in `dir` that are not directories as a `suti.Template`
// by calling `suti.LoadTemplateFile`. Partials for each template will be parsed from all
// files in a directory matching the base filename of the template (not including
// extension) if it exists.
func LoadTemplateDir(dir string) ([]suti.Template, error) {
	paths := make(map[string][]string) // [template][]partials

	if tpaths, err := loadPaths(dir); err != nil && !os.IsNotExist(err) {
		return nil, err
	} else {
		err = nil
		for _, t := range tpaths {
			if filepath.Ext(t) == ".ignore" {
				continue
			}
			paths[t] = make([]string, 0)
			dir, file := filepath.Split(t)
			ppath := filepath.Join(dir, strings.TrimSuffix(file, filepath.Ext(file)))
			for _, p := range tpaths {
				if !strings.Contains(p, ".ignore") && strings.Contains(p, ppath) && p != t {
					paths[t] = append(paths[t], p)
				}
			}
		}
	}

	var ret []suti.Template
	for t, partials := range paths {
		tmpl, err := suti.LoadTemplateFile(t, partials...)
		if err != nil {
			return nil, err
		}
		ret = append(ret, tmpl)
	}

	return ret, nil
}
