package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var templates = map[string]string{ // [ext]template
	"tmpl": "{{.Contents}}",
	"hmpl": "{{.Contents}}",
	"mst":  "{{Contents}}",
}

func createTestTemplates(dir string) (err error) {
	writef := func(path, data string) {
		if err == nil {
			err = os.WriteFile(path, []byte(data), 0644)
		}
	}

	for ext, data := range templates {
		writef(fmt.Sprintf("%s/root.%s", dir, ext), data)
		writef(fmt.Sprintf("%s/root.ignore.%s", dir, ext), data)
		writef(fmt.Sprintf("%s/root.%s.ignore", dir, ext), data)

		pdir := filepath.Join(dir, ext)
		err = os.Mkdir(pdir, 0755)
		writef(fmt.Sprintf("%s/partial.%s", pdir, ext), data)
		writef(fmt.Sprintf("%s/partial.ignore.%s", pdir, ext), data)
		writef(fmt.Sprintf("%s/partial.%s.ignore", pdir, ext), data)

		if err != nil {
			break
		}
	}

	return err
}

// file contents used in "contents" files
const contentsTxt = `p1
p2

	pre1
	pre2
p3

p4
`
const contentsHtml = `<p>p1<br>
p2</p>
<pre>pre1
pre2
</pre>
<p>p3</p>
<p>p4</p>`
const contentsMd = `p1
p2

	pre1
	pre2

p3
`
const contentsGfm = `p1
p2

	pre1
	pre2

p3`
const contentsCm = `p1
p2

	pre1
	pre2

p3`

var contents = map[string]string{
	".txt":  contentsTxt,
	".html": contentsHtml,
	".md":   contentsMd,
	".gfm":  contentsGfm,
	".cm":   contentsCm,
}

var asset = []byte{ // 5x5 black png image
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00,
	0x0d, 0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00,
	0x00, 0x05, 0x01, 0x03, 0x00, 0x00, 0x00, 0xb7, 0xa1, 0xb4, 0xa6,
	0x00, 0x00, 0x00, 0x06, 0x50, 0x4c, 0x54, 0x45, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0xa5, 0x67, 0xb9, 0xcf, 0x00, 0x00, 0x00, 0x02,
	0x74, 0x52, 0x4e, 0x53, 0xff, 0x00, 0xe5, 0xb7, 0x30, 0x4a, 0x00,
	0x00, 0x00, 0x09, 0x70, 0x48, 0x59, 0x73, 0x00, 0x00, 0x0e, 0xc4,
	0x00, 0x00, 0x0e, 0xc4, 0x01, 0x95, 0x2b, 0x0e, 0x1b, 0x00, 0x00,
	0x00, 0x10, 0x49, 0x44, 0x41, 0x54, 0x08, 0x99, 0x63, 0x60, 0x66,
	0x60, 0x66, 0x60, 0x00, 0x62, 0x76, 0x00, 0x00, 0x4a, 0x00, 0x11,
	0x3a, 0x34, 0x8c, 0xad, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
	0x44, 0xae, 0x42, 0x60, 0x82,
}

func createTestContents(dir string) (err error) {
	writef := func(path, data string) {
		if err == nil {
			err = os.WriteFile(path, []byte(data), 0644)
		}
	}

	for l, lang := range contentExts {
		if l == 0 {
			writef(fmt.Sprintf("%s/defaults.json", dir), "{ \"default\": \"data\" }")
		} else if l > 1 {
			dir, err = os.MkdirTemp(dir, "page")
		}
		writef(fmt.Sprintf("%s/.page.toml", dir), "page = \"data\"")
		writef(fmt.Sprintf("%s/body%d%s", dir, l, lang), contents[lang])
		writef(fmt.Sprintf("%s/asset.png", dir), string(asset))

		if err != nil {
			break
		}
	}

	return
}

func validateTestPagesNav(t *testing.T, pages []Page) {
	for _, p := range pages {
		var allUnique []string
		for _, navp := range p.Nav.All {
			for _, a := range allUnique {
				if a == navp.Path {
					t.Errorf("'%s' has .Nav.All items with duplicate .Path values (%s)",
						p.Path, a)
				}
			}
			allUnique = append(allUnique, navp.Path)
		}
		if len(p.Nav.All) != len(pages) {
			t.Errorf("'%s' has %d in .Nav.All (should be %d)",
				p.Path, len(p.Nav.All), len(pages))
		}

		foundAll := 0
		for _, navp := range p.Nav.All {
			for _, pp := range pages {
				if navp.Path == pp.Path {
					foundAll++
					break
				}
			}
		}
		if foundAll != len(p.Nav.All) {
			t.Errorf("found %d/%d pages in .Nav.All for '%s'",
				foundAll, len(p.Nav.All), p.Path)
		}

		foundRoot := false
		foundParent := false
		for _, pp := range pages {
			if !foundRoot {
				if p.Nav.Root == nil {
					foundRoot = true
				} else if p.Nav.Root.Path == pp.Path {
					foundRoot = true
				}
			}
			if !foundParent {
				if p.Nav.Parent == nil {
					foundParent = true
				} else if p.Nav.Parent.Path == pp.Path {
					foundParent = true
				}
			}
			if foundRoot && foundParent {
				break
			}
		}
		if !foundRoot {
			t.Errorf("could not find .Root '%s' for '%s'",
				p.Nav.Root.Path, p.Path)
		}
		if !foundParent {
			t.Errorf("could not find .Parent '%s' for '%s'",
				p.Nav.Parent.Path, p.Path)
		}

		// TODO test .Nav.Children, figure out how many should exist
		// TODO test .Nav.Crumbs, figure out how many should exist
	}
}

func validateTestPages(t *testing.T, pages []Page, e error) {
	if len(pages) != len(contents) {
		t.Fatalf("invalid number of pages returned (%d should be %d)",
			len(pages), len(contents))
	}

	var last, pt time.Time
	for i, p := range pages {
		if len(p.Slug) == 0 && p.Slug != filepath.Base(p.Path) {
			t.Errorf("bad Slug for page: '%s' (%s) - should be '%s'",
				p.Slug, p.Path, filepath.Base(p.Path))
		}
		if len(p.Path) == 0 {
			t.Error("empty Path for page:", p)
		}
		validateTestPagesNav(t, pages)
		if _, ok := p.Meta["page"]; !ok || len(p.Meta) == 0 {
			t.Errorf("missing page Meta key for page: '%s'", p.Path)
		}
		if _, ok := p.Meta["default"]; !ok || len(p.Meta) == 0 {
			t.Error("empty default Meta key for page:", p.Path)
		}
		if len(p.Contents) == 0 {
			t.Error("empty Contents for page:", p.Path)
		}
		if len(p.Assets) == 0 {
			t.Error("empty Assets for page:", p.Path)
		}
		if pt, e = time.Parse(timefmt, p.Updated); e != nil {
			t.Fatal(e)
		}

		if i == 0 {
			last = pt
		} else if pt.Before(last) {
			for _, pp := range pages {
				t.Logf("%s - %s", pp.Path, pp.Updated)
			}
			t.Error("Contents Pages returned in wrong order")
		}
	}
}
