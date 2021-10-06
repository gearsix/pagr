package main

import (
	"fmt"
	"path/filepath"
	"os"
	"time"
	"notabug.org/gearsix/suti"
	"testing"
)

// helper functions
func validateContents(t *testing.T, pages []Page, e error) {
	if len(pages) != len(contents)-1 {
		t.Fatalf("invalid number of pages returned (%d should be %d)",
			len(pages), len(contents))
	}

	var last, pt time.Time
	for i, p := range pages {
		if len(p.Slug) == 0 && p.Slug != filepath.Base(p.Path) {
			t.Errorf("empty Slug for page: '%s'. Should be '%s'", p.Slug, filepath.Base(p.Path))
		}
		if len(p.Path) == 0 {
			t.Error("empty Path for page:", p)
		}
		// TODO test p.Nav here
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
				t.Log(pp.Updated)
			}
			t.Error("Contents Pages returned in wrong order")
		}
	}
}

var contents = map[string]string{
	".txt": `p1
p2

	pre1
	pre2
p3

p4
`,
	".html": `<p>p1<br>
p2</p>
<pre>pre1
pre2
</pre>
<p>p3</p>
<p>p4</p>`,
	".md": `p1
p2

	pre1
	pre2

p3
`,
	".gfm": `p1
p2

	pre1
	pre2

p3`,
	".cm": `p1
p2

	pre1
	pre2

p3`,
}

var asset = []byte{ // 5x5 black png
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

func createProjectContents(dir string) (err error) {
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

func TestBuildSitemap(test *testing.T) {
	test.Parallel()

	var err error
	/*
	writef := func(path, data string) {
		if err == nil {
			err = os.WriteFile(path, []byte(data), 0644)
		}
	}
	*/
	tdir := test.TempDir()
	// TODO write files to pages dir

	var p []Page
	if p, err = LoadPagesDir(tdir); err != nil {
		test.Errorf("LoadPagesDir failed: %s", err)
	}
	p = BuildSitemap(p)
	// TODO validate p
}

func TestLoadPagesDir(test *testing.T) {
	var err error
	tdir := test.TempDir()
	if err = createProjectContents(tdir); err != nil {
		test.Errorf("failed to create test content: %s", err)
	}

	var p []Page
	if p, err = LoadPagesDir(tdir); err != nil {
		test.Fatalf("LoadPagesDir failed: %s", err)
	}

	validateContents(test, p, err)
}

func TestMergeMeta(test *testing.T) {
	test.Parallel()

	var orig Meta
	merge := make(Meta)
	merge["test"] = "overwritten"
	merge["new"] = "data"

	orig = make(Meta)
	orig["test"] = "data"
	orig.MergeMeta(merge, false)
	if val, ok := orig["test"]; ok {
		if val == "overwritten" {
			test.Fatalf("invalid 'test' value: %s", val)
		}
	} else if !ok {
		test.Fatalf("unable to parse 'test' key")
	}
	if _, ok := orig["new"]; !ok {
		test.Fatalf("unable to parse 'new' key")
	}

	orig = make(Meta)
	orig["test"] = "data"
	orig.MergeMeta(merge, true)
	if val, ok := orig["test"]; ok {
		if val != "overwritten" {
			test.Fatalf("invalid 'test' value: %s", val)
		}
	} else if !ok {
		test.Fatalf("unable to parse 'test' key")
	}
	if _, ok := orig["new"]; !ok {
		test.Fatalf("unable to parse 'new' key")
	}
}

func TestNewPage(test *testing.T) {
	test.Parallel()

	const path = "/test/path"
	var updated = time.Now()

	p := NewPage(path, updated)

	if p.Slug != "path" || p.Path != path || p.Updated != updated.Format(timefmt) {
		test.Fatal("invalid Page", p)
	}
}

func TestTemplateName(test *testing.T) {
	test.Parallel()

	p := NewPage("/test", time.Now())
	if p.TemplateName() != DefaultTemplateName {
		test.Fatalf("'%s' not returned from TemplateName()", DefaultTemplateName)
	}
	p.Meta["Template"] = "test1"
	if p.TemplateName() != "test1" {
		test.Fatalf("'test1' not returned from TemplateName()")
	}
	p.Meta["template"] = "test2"
	if p.TemplateName() != "test2" {
		test.Fatalf("'test2' not returned from TemplateName()")
	}
}

func TestNewContentFromFile(test *testing.T) {
	test.Parallel()

	var err error
	contents := map[string]string {
		"txt": `test`,
		"md": "**test**\ntest",
		"gfm": "**test**\ntest",
		"cm": "**test**",
		"html": `<b>test</b>`,
	}

	tdir := test.TempDir()
	contentsPath := func(ftype string) string {
		return tdir + "/test." + ftype
	}

	for ftype, data := range contents {
		if err = os.WriteFile(contentsPath(ftype), []byte(data), 0666); err != nil {
			test.Error("TestNewContentFromFile setup failed:", err)
		}
	}

	var p Page
	for ftype := range contents {
		if err = p.NewContentFromFile(contentsPath(ftype)); err != nil {
			test.Fatal("NewContentFromFile failed for", ftype, err)
		}
	}
}

func TestCopyFile(test *testing.T) {
	test.Parallel()

	var err error
	tdir := test.TempDir()
	src := tdir + "/src"
	dst := tdir + "/dst"
	data := []byte("data")
	if err = os.WriteFile(src, data, 0666); err != nil {
		test.Error("setup failed, could not write", tdir+"/src")
	}

	if err = CopyFile(src, dst); err != nil {
		test.Fatal("CopyFile failed", err)
	}
	if _, err = os.Stat(dst); err != nil {
		test.Fatalf("could not stat '%s'", dst)
	}
	var buf []byte
	if buf, err = os.ReadFile(dst); err != nil {
		test.Errorf("could not read '%s'", dst)
	} else if len(buf) < len(data) {
		test.Fatalf("not all data (%s) copied to '%s' (%s)", data, dst, buf)
	} else if string(buf) != string(data) {
		test.Fatalf("copied data (%s) does not match source (%s)", buf, data)
	}
}

func TestCopyAssets(test *testing.T) {
	test.Parallel()

	var p Page
	var err error

	srcDir := test.TempDir()
	src := []string{"1","2","3","4"}

	for _, fname := range src {
		p.Assets = append(p.Assets, fname)
		path := filepath.Join(srcDir, fname)
		if _, err = os.Create(path); err != nil {
			test.Fatalf("failed to create source file '%s'", path)
		}
	}

	dstDir := test.TempDir()
	if err = p.CopyAssets(srcDir, dstDir); err != nil {
		test.Fatal("CopyAssets failed", err)
	}
	for _, fname := range src {
		if _, err := os.Stat(dstDir+"/"+fname); err != nil {
			test.Fatal("missing file", dstDir+"/"+fname)
		}
	}
}

func TestBuild(test *testing.T) {
	test.Parallel()

	// setup
	var err error
	tdir := test.TempDir()
	p := NewPage("/test", time.Now())
	t, err := suti.LoadTemplateString("tmpl", "test", `{{.Meta.Title}} {{template "p" .}}`, map[string]string{"p": "p"})
	if err != nil {
		test.Error(err)
	}

	var fpath string
	if fpath, err = p.Build(tdir, t); err != nil {
		test.Fatal(err)
	}
	var fbuf []byte
	if fbuf, err = os.ReadFile(fpath); err != nil {
		test.Fatal(err)
	}
	if string(fbuf) != "Test p" {
		test.Fatalf("invalid result parsed: '%s', expected: 'Test p'", string(fbuf))
	}
}

