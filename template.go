package main

import (
	"path/filepath"
	"io/fs"
	"strings"
	"os"
	"notabug.org/gearsix/suti"
)

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

func LoadTemplateDir(dir string) ([]suti.Template, error) {
	paths := make(map[string][]string) // [template][]partials

	if tpaths, err := loadPaths(dir); err != nil && !os.IsNotExist(err) {
		return nil, err
	} else {
		err = nil
		for _, t := range tpaths {
			paths[t] = make([]string, 0)
			dir, file := filepath.Split(t)
			ppath := filepath.Join(dir, strings.TrimSuffix(file, filepath.Ext(file)))
			if _, err := os.Stat(ppath); err == nil {
				var partials []string
				if partials, err = loadPaths(ppath); err != nil {
					return nil, err
				} else {
					paths[t] = partials
				}
			}
		}
	}

	var err error
	var ret []suti.Template
	for t, partials := range paths {
		tmpl, err := suti.LoadTemplateFile(t, partials...)
		if err != nil {
			break
		}
		ret = append(ret, tmpl)
	}

	return ret, err
}
