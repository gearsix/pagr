package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/yuin/goldmark"
	goldmarkext "github.com/yuin/goldmark/extension"
	goldmarkparse "github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	"io"
	"io/fs"
	"time"
	"notabug.org/gearsix/suti"
	"os"
	"path/filepath"
	"strings"
	"sort"
)

// SupportedContent provides a list of supported file extensions for Content
// files. Any file in the Content directory not matching one of these will be
// ignored unless it's a Meta file.
var SupportedContent = [5]string{
	".txt",  // plain-text
	".html", // HTML
	".md",   // commonmark + extensions (linkify, auto-heading id, unsafe HTML)
	".gfm",  // github-flavoured markdown
	".cm",   // commonmark
}

func titleFromPath(path string) (title string) {
	if title = filepath.Base(path); title == "/" {
		title = "Home"
	}
	title = strings.TrimSuffix(title, filepath.Ext(title))
	title = strings.ReplaceAll(title, "-", " ")
	title = strings.Title(title)
	return
}

func lastFileMod(fpath string) time.Time {
	t := time.Now() // default/error ret
	if fd, e := os.Stat(fpath); e != nil {
		return t
	} else if !fd.IsDir() {
		return fd.ModTime()
	} else {
		t = fd.ModTime()
	}
	if dir, err := os.ReadDir(fpath); err != nil {
		return t
	} else {
		for _, d := range dir {
			if fd, err := d.Info(); err == nil && fd.ModTime().After(t) {
				t = fd.ModTime()
			}
		}
	}
	return t
}

type Content []Page

// Sitemap parses `pages` to determine the `.PageNav` values for each element in `pages`
// based on their `.Path` value. These values will be set in the returned Content
func BuildSitemap(pages Content) Content {
	var root *Page
	for i, p := range pages {
		if p.Path == "/" {
			root = &pages[i]
			break
		}
	}

	for i, p := range pages {
		p.Nav.Root = root

		pdepth := len(strings.Split(p.Path, "/")[1:])
		if p.Path == "/" {
			pdepth = 0
		}

		if pdepth == 1 && p.Path != "/" {
			p.Nav.Parent = root
		}

		for j, pp := range pages {
			ppdepth := len(strings.Split(pp.Path, "/")[1:])
			if pp.Path == "/" {
				ppdepth = 0
			}

			p.Nav.All = append(p.Nav.All, &pages[j])
			if p.Nav.Parent == nil && ppdepth == pdepth - 1 && strings.Contains(p.Path, pp.Path) {
				p.Nav.Parent = &pages[j]
			}
			if ppdepth == pdepth + 1 && strings.Contains(pp.Path, p.Path) {
				p.Nav.Children = append(p.Nav.Children, &pages[j])
			}
		}

		var crumb string
		for _, c := range strings.Split(p.Path, "/")[1:] {
			crumb += "/" + c
			for j, pp := range pages {
				if pp.Path == crumb {
					p.Nav.Crumbs = append(p.Nav.Crumbs, &pages[j])
					break
				}
			}
		}

		pages[i] = p
	}

	return pages
}

// LoadContentDir parses all files/directories in `dir` into a `Content`.
// For each directory, a new `Page` element will be generated, any file with a
// filetype found in `SupportedContent`, will be parsed into a string of HTML
// and appended to the `.Content` of the `Page` generated for it's parent
// directory.
func LoadContentDir(dir string) (c Content, e error) {
	if _, e = os.Stat(dir); e != nil {
		return
	}

	pages := make(map[string]Page)
	defaults := make(map[string]Meta)
	if dir[len(dir)-1] != '/' {
		dir += "/"
	}
	e = filepath.Walk(dir, func(fpath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.Contains(fpath, ".ignore") {
			return nil
		}

		var path string
		if info.IsDir() {
			path = "/" + strings.TrimPrefix(fpath, dir)
			page := NewPage(path)
			for i, p := range path {
				if p != '/' {
					continue
				}
				dpath := path[:i]
				if len(dpath) == 0 {
					dpath = "/"
				}
				if _, ok := defaults[dpath]; ok {
					page.Meta.MergeMeta(defaults[dpath], true)
				}
			}
			pages[path] = page
			return nil
		}

		path, _ = filepath.Split(fpath)
		path = strings.TrimPrefix(path, dir)
		path = "/" + strings.TrimSuffix(path, "/")
		page := pages[path]

		if strings.Contains(fpath, ".page") || strings.Contains(fpath, ".default") {
			var m Meta
			if err = suti.LoadDataFile(fpath, &m); err != nil {
				return err
			}
			if strings.Contains(fpath, ".page") {
				page.Meta.MergeMeta(m, true)
			} else if strings.Contains(fpath, ".defaults") {
				defaults[path] = m
			}
		} else if isSupportedContentExt(filepath.Ext(fpath)) > -1 {
			err = page.NewContentFromFile(fpath)
		} else {
			page.Assets = append(page.Assets, fpath)
		}

		pages[path] = page
		return nil
	})

	for _, page := range pages {
		c = append(c, page)
	}
	sort.SliceStable(c, func(i, j int) bool { return c[i].Updated.Before(c[j].Updated) })

	c = BuildSitemap(c)

	return c, e
}

func isSupportedContentExt(ext string) int {
	for i, supported := range SupportedContent {
		if ext == supported {
			return i
		}
	}
	return -1
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

// Page is the data structure loaded from Content files/folders that
// gets passed to templates for execution after Content has been loaded.
// This is the data structure to reference when writing a template!
type Page struct {
	Title    string
	Path     string
	Nav      PageNav
	Meta     Meta
	Contents []string
	Assets   []string
	Updated  time.Time
}

// PageNav is a struct that provides a set of pointers for navigating a
// across a set of pages. All values are initialised to nil and will only
// be populated manually or by calling `BuildSitemap`.
type PageNav struct {
	All      []*Page
	Root     *Page
	Parent   *Page
	Children []*Page
	Crumbs   []*Page
}

// NewPage returns a Page with init values. `.Title` will be set to the
// value returned by titleFromPath(path), `.Path` will be set to `path`.
// Updated is set to time.Now(). Any other values will simply be initialised.
func NewPage(path string) Page {
	return Page{
		Title:    titleFromPath(path),
		Path:     path,
		Nav:      PageNav{},
		Meta:     make(Meta),
		Contents: make([]string, 0),
		Assets:   make([]string, 0),
		Updated:  lastFileMod(path),
	}
}

// GetTemplate will check if `p.Meta` has the key `template` or `Template`
// (in the order) and return the value of the first existing key as a string.
// If `.Meta` neither has the key `template` or `Template`, then it will
// return `DefaultTemplate` from [./template.go].
func (p *Page) GetTemplate() string {
	if v, ok := p.Meta["template"]; ok {
		return v.(string)
	} else if v, ok = p.Meta["Template"]; ok {
		return v.(string)
	} else {
		return DefaultTemplate
	}
}

// NewContentFromFile loads the file from `fpath` and converts it to HTML
// from the language matching it's file extension (see below).
// - ".txt" = plain-text
// - ".md", ".gfm", ".cm" = various flavours of markdown
// - ".html" = parsed as-is
// Successful conversions are appended to `p.Contents`
func (p *Page) NewContentFromFile(fpath string) (err error) {
	var buf []byte
	if f, err := os.Open(fpath); err == nil {
		buf, err = io.ReadAll(f)
		f.Close()
	}
	if err != nil {
		return
	}

	var body string
	for _, lang := range SupportedContent {
		if filepath.Ext(fpath) == lang {
			switch lang {
			case ".txt":
				body = convertTextToHTML(bytes.NewReader(buf))
			case ".md":
				fallthrough
			case ".gfm":
				fallthrough
			case ".cm":
				body, err = convertMarkdownToHTML(lang, buf)
			case ".html":
				body = string(buf)
			default:
				break
			}
		}
	}
	if len(body) == 0 {
		err = fmt.Errorf("invalid filetype (%s) passed to NewContentFromFile",
			filepath.Ext(fpath))
	}
	if err == nil {
		p.Contents = append(p.Contents, body)
	}

	return err
}

// convertTextToHTML parses textual data from `in` and line-by-line converts
// it to HTML. Conversion rules are as follows:
// - Blank lines (with escape characters trimmed) will close any opon tags
// - If a text line is prefixed with a tab and no tag is open, it will open a <pre> tag
// - Otherwise any line of text will open a <p> tag
func convertTextToHTML(in io.Reader) (html string) {
	var tag int
	const p = 1
	const pre = 2

	fscan := bufio.NewScanner(in)
	for fscan.Scan() {
		line := fscan.Text()
		if len(strings.TrimSpace(line)) == 0 {
			switch tag {
			case p:
				html += "</p>\n"
			case pre:
				html += "</pre>\n"
			}
			tag = 0
		} else if tag == 0 && line[0] == '\t' {
			tag = pre
			html += "<pre>" + line[1:] + "\n"
		} else if tag == 0 || (tag == pre && line[0] != '\t') {
			if tag == pre {
				html += "</pre>\n"
			}
			tag = p
			html += "<p>" + line
		} else if tag == p {
			html += " " + line
		} else if tag == pre {
			html += line[1:] + "\n"
		}
	}
	if tag == p {
		html += "</p>"
	} else if tag == pre {
		html += "</pre>"
	}

	return html
}

// convertMarkdownToHTML initialises a `goldmark.Markdown` based on `lang` and
// returns values from calling it's `Convert` function on `in`.
// Markdown `lang` options, see the code for specfics:
// - ".gfm" = github-flavoured markdown
// - ".cm" = standard commonmark
// - ".md" (and anything else) = commonmark + extensions (linkify, auto-heading id, unsafe HTML)
func convertMarkdownToHTML(lang string, buf []byte) (string, error) {
	var markdown goldmark.Markdown
	switch lang {
	case ".gfm":
		markdown = goldmark.New(
			goldmark.WithExtensions(
				goldmarkext.GFM,
				goldmarkext.Table,
				goldmarkext.Strikethrough,
				goldmarkext.Linkify,
				goldmarkext.TaskList,
			),
			goldmark.WithParserOptions(
				goldmarkparse.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				goldmarkhtml.WithUnsafe(),
				goldmarkhtml.WithHardWraps(),
			),
		)
	case ".cm":
		markdown = goldmark.New()
	case ".md":
		fallthrough
	default:
		markdown = goldmark.New(
			goldmark.WithExtensions(
				goldmarkext.Linkify,
			),
			goldmark.WithParserOptions(
				goldmarkparse.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				goldmarkhtml.WithUnsafe(),
			),
		)
	}

	var out bytes.Buffer
	err := markdown.Convert(buf, &out)
	return out.String(), err
}
