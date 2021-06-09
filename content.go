package main

import (
    "bytes"
    "bufio"
    "path/filepath"
    "io"
    "io/fs"
    "strings"
    "os"
    "notabug.org/gearsix/suti"
	"github.com/yuin/goldmark"
	goldmarkext "github.com/yuin/goldmark/extension"
	goldmarkparse "github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
)

var SupportedContentFiletypes = []string{
    ".txt",  // plain-text
    ".html", // HTML
    ".md",   // commonmark with non-intrusive extensions: linkify, auto heading id, unsafe HTML
    ".gfm",  // github-flavoured markdown
    ".cm",   // commonmark
}

type Content []Page

func LoadContentDir(path string) (c Content, e error) {
    pages := make(map[string]Page)
    defaults := make(map[string]Meta)
    e = filepath.Walk(path, func(fpath string, info fs.FileInfo, err error) error {
        if err != nil {
            return nil
        }

        if info.IsDir() {
            p := NewPage(strings.TrimPrefix(fpath, path))
            for _, dir := range strings.Split(fpath, "/") {
                if _, ok := defaults[dir]; ok {
                    p.Meta.MergeMeta(defaults[dir], true)
                }
            }
            return nil
        }

        pdir := filepath.Dir(fpath)
        page := pages[pdir]
        if strings.Contains(fpath, ".page") || strings.Contains(fpath, ".defaults") {
            var m Meta
            if err = suti.LoadDataFile(fpath, &m); err != nil {
                return err
            }
            if strings.Contains(fpath, ".page") {
                page.Meta.MergeMeta(m, true)
            } else if strings.Contains(fpath, ".defaults") {
                defaults[pdir] = m
            }
        } else if ext := filepath.Ext(fpath); ext == ".txt" || ext == ".md" || ext == ".html" {
            page.NewBodyFromFile(fpath)
        } else {
            page.Assets = append(page.Assets, strings.TrimPrefix(fpath, path))
        }

        pages[pdir] = page
        return nil
    })

    for _, page := range pages {
        c = append(c, page)
    }

    return c, e
}

type Meta map[string]interface{}

func (m Meta) MergeMeta(meta Meta, overwrite bool) {
    for k, v := range meta {
        if _, ok := m[k]; ok && overwrite {
            m[k] = v
        } else if !ok {
            m[k] = v
        }
    }
}

type Page struct {
    Path string
    Meta Meta
    Body []string
    Assets []string
}

func NewPage(path string) Page {
    return Page {
        Path: path,
        Meta: make(Meta),
        Body: make([]string, 0),
        Assets: make([]string, 0),
    }
}

func (p *Page) NewBodyFromFile(fpath string) (err error) {
    var buf []byte
    if f, err := os.Open(fpath); err == nil {
        buf, err = io.ReadAll(f)
        f.Close()
    }
    if err != nil {
        return
    }

    var body string
    for _, lang := range SupportedContentFiletypes {
        if filepath.Ext(fpath) == lang {
            switch (lang) {
            case ".txt":
                body = txt2html(bytes.NewReader(buf))
            case ".md":
            case ".gmd":
            case ".cm":
                markdown := getMarkdown(lang)
                var out bytes.Buffer
                if err = markdown.Convert(buf, &out); err == nil {
                    body = out.String()
                }
            case ".html":
                body = string(buf)
            default:
                continue
            }
        }
    }

    if len(body) == 0 {
        panic("passed invalid filetype to NewBodyFromFile")
    }
    p.Body = append(p.Body, body)

    return err
}

// txt2html parses textual data from `in` and line-by-line converts
// it to HTML. Conversion rules are as follows:
// - Blank lines (with escape characters trimmed) will close any opon tags
// - If a text line is prefixed with a tab and no tag is open, it will open a <pre> tag
// - Otherwise any line of text will open a <p> tag
func txt2html(in io.Reader) (html string) {
	var block int
	const p = 1
	const pre = 2

	fscan := bufio.NewScanner(in)
	for fscan.Scan() {
		line := fscan.Text()
		if len(strings.TrimSpace(line)) == 0 {
			switch block {
			case p:
				html += "</p>\n"
			case pre:
				html += "</pre>\n"
			}
			block = 0
		} else if block == 0 && line[0] == '\t' {
			block = pre
			html += "<pre>" + line + "\n"
		} else if block == 0 {
			block = p
			html += "<p>" + line
		} else if block == p {
			html += " " + line
		} else if block == pre {
			html += line + "\n"
		}
	}
	if block == p {
		html += "</p>"
	} else if block == pre {
		html += "</pre>"
	}

	return html
}

func getMarkdown(lang string) (markdown goldmark.Markdown) {
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
    return
}

