package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	goldmarkext "github.com/yuin/goldmark/extension"
	goldmarkparse "github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	"notabug.org/gearsix/suti"
)

// Content is the converted HTML string of a Content file
type Content string

var contentExts = [4]string{
	"",      // pre-formatted text
	".txt",  // plain-text
	".html", // HTML
	".md",   // commonmark + extensions (linkify, auto-heading id, unsafe HTML)
}

func isContentExt(ext string) int {
	for i, supported := range contentExts {
		if ext == supported {
			return i
		}
	}
	return -1
}

// FIX kills performance on windows
func gitModTime(fpath string) (mod time.Time, err error) {
	if gitBin == "" {
		err = fmt.Errorf("git binary not found")
		return
	}

	if fpath, err = filepath.Abs(fpath); err != nil {
		return
	}

	git := exec.Command(gitBin, "-C", filepath.Dir(fpath), "log", "-1", "--format='%ad'", "--", fpath)
	var out []byte
	if out, err = git.Output(); err == nil {
		outstr := strings.ReplaceAll(string(out), "'", "")
		outstr = strings.TrimSuffix(outstr, "\n")
		mod, err = time.Parse("Mon Jan 2 15:04:05 2006 -0700", outstr)
	}
	return
}

func lastPageMod(fpath string) (t time.Time) {
	if fd, err := os.Stat(fpath); err != nil {
		if t, err = gitModTime(fpath); err != nil {
			t = time.Now()
		}
	} else {
		if t, err = gitModTime(fpath); err != nil {
			t = fd.ModTime()
		}

		if fd.IsDir() { // find last modified file in directory (depth 1)
			var dir []os.FileInfo
			if dir, err = ioutil.ReadDir(fpath); err == nil {
				for i, f := range dir {
					if f.IsDir() {
						continue
					}

					var ft time.Time
					if ft, err = gitModTime(filepath.Join(fpath, f.Name())); err != nil {
						ft = fd.ModTime()
					}

					if i == 0 || ft.After(t) {
						t = ft
					}
				}
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
func LoadContentDir(dir string) (p []Page, e error) {
	if _, e = os.Stat(dir); e != nil {
		return
	}
	dir = filepath.Clean(dir)

	pages := make(map[string]Page)
	dmeta := make(map[string]Meta)

	e = filepath.Walk(dir, func(fpath string, info os.FileInfo, err error) error {
		if err != nil || ignoreFile(fpath) {
			return err
		}

		if info.IsDir() {
			path := pagePath(dir, fpath)
			pages[path] = NewPage(path, lastPageMod(fpath))
		} else {
			path := pagePath(dir, filepath.Dir(fpath))
			pages[path], dmeta, err = loadContentFile(pages[path], dmeta, fpath, path)
		}
		return err
	})

	for _, page := range pages {
		page.applyDefaults(dmeta)
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

func loadContentFile(p Page, def map[string]Meta, fpath string, ppath string) (Page, map[string]Meta, error) {
	var err error
	fname := strings.TrimSuffix(filepath.Base(fpath), filepath.Ext(fpath))

	if suti.IsSupportedDataLang(filepath.Ext(fpath)) != -1 &&
		(fname == "defaults" || fname == "meta") {
		var m Meta
		if err = suti.LoadDataFilepath(fpath, &m); err == nil {
			if fname == "defaults" || fname == "default" {
				if meta, ok := def[ppath]; ok {
					m.MergeMeta(meta, false)
				}
				def[ppath] = m
			} else if fname == "meta" {
				p.Meta.MergeMeta(m, true)
			}
		}
	} else if isContentExt(filepath.Ext(fpath)) != -1 {
		err = p.NewContentFromFile(fpath)
	} else {
		a := filepath.Join(ppath, filepath.Base(fpath))
		p.Assets.All = append(p.Assets.All, a)
		ref := &p.Assets.All[len(p.Assets.All)-1]
		mimetype := mime.TypeByExtension(filepath.Ext(fpath))
		if strings.Contains(mimetype, "image") {
			p.Assets.Image = append(p.Assets.Image, ref)
		} else if strings.Contains(mimetype, "video") {
			p.Assets.Video = append(p.Assets.Video, ref)
		} else if strings.Contains(mimetype, "audio") {
			p.Assets.Audio = append(p.Assets.Audio, ref)
		} else {
			p.Assets.Misc = append(p.Assets.Misc, ref)
		}
	}
	return p, def, err
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
		buf, err = ioutil.ReadAll(f)
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
				body, err = convertMarkdownToHTML(buf)
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
// ".md" (and anything else) = commonmark + extensions (linkify, auto-heading id, unsafe HTML)
func convertMarkdownToHTML(buf []byte) (md string, err error) {
	markdown := goldmark.New(
		goldmark.WithExtensions(goldmarkext.Linkify),
		goldmark.WithParserOptions(goldmarkparse.WithAutoHeadingID()),
		goldmark.WithRendererOptions(goldmarkhtml.WithUnsafe()),
	)
	var out bytes.Buffer
	err = markdown.Convert(buf, &out)
	return out.String(), err
}
