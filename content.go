package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"github.com/yuin/goldmark"
	goldmarkext "github.com/yuin/goldmark/extension"
	goldmarkparse "github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	"notabug.org/gearsix/suti"
	"path/filepath"
	"os"
	"strings"
	"sort"
	"time"
)

type Content string

var contentExts = [6]string{
	"",      // pre-formatted text
	".txt",  // plain-text
	".html", // HTML
	".md",   // commonmark + extensions (linkify, auto-heading id, unsafe HTML)
	".gfm",  // github-flavoured markdown
	".cm",   // commonmark
}

func isContentExt(ext string) int {
	for i, supported := range contentExts {
		if ext == supported {
			return i
		}
	}
	return -1
}

func lastModFile(fpath string) (t time.Time) {
	if fd, e := os.Stat(fpath); e != nil {
		t = time.Now()
	} else if !fd.IsDir() {
		t = fd.ModTime()
	} else { // find last modified file in directory (depth 1)
		dir, err := os.ReadDir(fpath)
		if err != nil {
			return t
		}

		for i, d := range dir {
			if fd, err := d.Info(); err == nil && (i == 0 || fd.ModTime().After(t)) {
				t = fd.ModTime()
			}
		}
	}
	return
}

// LoadContentsDir parses all files/directories in `dir` into a `Content`.
// For each directory, a new `Page` element will be generated, any file with a
// filetype found in `contentExts`, will be parsed into a string of HTML
// and appended to the `.Content` of the `Page` generated for it's parent
// directory.
func LoadContentsDir(dir string) (p []Page, e error) {
	if _, e = os.Stat(dir); e != nil {
		return
	}
	dir = filepath.Clean(dir)

	pages := make(map[string]Page)
	dmetas := make(map[string]Meta)

	e = filepath.Walk(dir, func(fpath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if ignoreFile(fpath) {
			return nil
		}

		if info.IsDir() {
			path := pagePath(dir, fpath)
			pages[path] = NewPage(path, lastModFile(fpath))
		} else {
			path := pagePath(dir, filepath.Dir(fpath))
			page := pages[path]

			if suti.IsSupportedDataLang(filepath.Ext(fpath)) > -1 {
				var m Meta
				if err = suti.LoadDataFilepath(fpath, &m); err == nil {
					if strings.Contains(filepath.Base(fpath), "defaults.") ||
						strings.Contains(filepath.Base(fpath), "default.") {
						if meta, ok := dmetas[path]; ok {
							m.MergeMeta(meta, false)
						}
						dmetas[path] = m
					} else {
						page.Meta.MergeMeta(m, true)
					}
				}
			} else if isContentExt(filepath.Ext(fpath)) > -1 {
				err = page.NewContentFromFile(fpath)
			} else if suti.IsSupportedDataLang(filepath.Ext(fpath)) == -1 {
				page.Assets = append(page.Assets, filepath.Join(path, filepath.Base(fpath)))
			}

			pages[path] = page
		}
		return err
	})

	for _, page := range pages {
		page.applyDefaults(dmetas)
		p = append(p, page)
	}

	sort.SliceStable(p, func(i, j int) bool {
		if it, err := time.Parse(timefmt, p[i].Updated); err == nil {
			if jt, err := time.Parse(timefmt, p[j].Updated); err == nil {
				return it.After(jt)
			}
		}
		return false
	})

	p = BuildSitemap(p)

	return
}

// NewContentFromFile loads the file from `fpath` and converts it to HTML
// from the language matching it's file extension (see below).
// - ".txt" = plain-text
// - ".md", ".gfm", ".cm" = various flavours of markdown
// - ".html" = parsed as-is
// Successful conversions are appended to `p.Contents`
func NewContentFromFile(fpath string) (c Content, err error) {
	var buf []byte
	if f, err := os.Open(fpath); err == nil {
		buf, err = io.ReadAll(f)
		f.Close()
	}
	if err != nil {
		return
	}

	var body string
	for _, lang := range contentExts {
		if filepath.Ext(fpath) == lang {
			switch lang {
			case "":
				body = "<pre>" + string(buf) + "</pre>"
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
	c = Content(body)
	return
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
func convertMarkdownToHTML(lang string, buf []byte) (md string, err error) {
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
	err = markdown.Convert(buf, &out)
	return out.String(), err
}
