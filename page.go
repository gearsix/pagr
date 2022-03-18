package main

import (
	"bytes"
	"io/ioutil"
	"notabug.org/gearsix/suti"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const timefmt = "2006-01-02"

func titleFromPath(path string) (title string) {
	if title = filepath.Base(path); title == "/" {
		title = "Home"
	}
	title = strings.TrimSuffix(title, filepath.Ext(title))
	title = strings.ReplaceAll(title, "-", " ")
	title = strings.Title(title)
	return
}

func pagePath(root, path string) string {
	path = strings.TrimPrefix(path, root)
	if len(path) == 0 {
		path = "/"
	} else {
		path = filepath.ToSlash(path)
	}
	return path
}

// Page is the data structure loaded from Content files/folders that
// gets passed to templates for execution after Content has been loaded.
// This is the data structure to reference when writing a template!
type Page struct {
	Slug     string
	Path     string
	Nav      Nav
	Meta     Meta
	Contents []Content
	Assets   []string
	Updated  string
}

// Nav is a struct that provides a set of pointers for navigating a
// across a set of pages. All values are initialised to nil and will only
// be populated manually or by calling `BuildSitemap`.
type Nav struct {
	All      []*Page
	Root     *Page
	Parent   *Page
	Children []*Page
	Crumbs   []*Page
}

// Meta is the structure any metadata is parsed into (_.toml_, _.json_, etc)
type Meta map[string]interface{}

// MergeMeta merges `meta` into `m`. When there are matching keys in both,
// `overwrite` determines whether the existing value in `m` is overwritten.
func (m Meta) MergeMeta(meta Meta, overwrite bool) {
	for k, v := range meta {
		if _, ok := m[k]; ok && overwrite {
			m[k] = v
		} else if !ok {
			m[k] = v
		}
	}
}

// NewPage returns a Page with init values. `.Path` will be set to `path`.
// Updated is set to time.Now(). Any other values will simply be initialised.
func NewPage(path string, updated time.Time) Page {
	return Page{
		Slug:     filepath.Base(path),
		Path:     path,
		Nav:      Nav{},
		Meta:     Meta{"Title": titleFromPath(path)},
		Contents: make([]Content, 0),
		Assets:   make([]string, 0),
		Updated:  updated.Format(timefmt),
	}
}

// GetTemplate will check if `p.Meta` has the key `template` or `Template`
// (in the order) and return the value of the first existing key as a string.
// If `.Meta` neither has the key `template` or `Template`, then it will
// return `DefaultTemplateName` from [./template.go].
func (p *Page) TemplateName() string {
	if v, ok := p.Meta["template"]; ok {
		return v.(string)
	} else if v, ok = p.Meta["Template"]; ok {
		return v.(string)
	} else {
		return DefaultTemplateName
	}
}

// Build will run `t.Execute(p)` and write the result to
// `outDir/p.Path/index.html`.
func (p *Page) Build(outDir string, t suti.Template) (out string, err error) {
	var buf bytes.Buffer
	if buf, err = t.Execute(p); err == nil {
		out = filepath.Join(outDir, p.Path, "index.html")
		if err = os.MkdirAll(filepath.Dir(out), 0755); err == nil {
			err = ioutil.WriteFile(out, buf.Bytes(), 0644)
		}
	}
	return out, err
}

// call `NewContentFromFile` and append it to `p.Contents`
func (p *Page) NewContentFromFile(fpath string) (err error) {
	var c Content
	if c, err = NewContentFromFile(fpath); err == nil {
		p.Contents = append(p.Contents, c)
	}
	return
}

func (page *Page) applyDefaults(defaultMetas map[string]Meta) {
	for i, p := range page.Path {
		if p != '/' {
			continue
		}
		path := page.Path[:i]
		if len(path) == 0 {
			path = "/"
		}
		if meta, ok := defaultMetas[path]; ok {
			page.Meta.MergeMeta(meta, false)
		}
	}
}
