package main

import (
	"io/ioutil"
	"notabug.org/gearsix/suti"
	"os"
	"path/filepath"
	"testing"
	"time"
)

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

func TestBuild(test *testing.T) {
	test.Parallel()

	var err error
	tdir := filepath.Join(os.TempDir(), "pagr_test", "TestBuild")
	if err := os.MkdirAll(tdir, 0775); err != nil {
		test.Errorf("failed to create temporary test dir: %s", tdir)
	}
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
	if fbuf, err = ioutil.ReadFile(fpath); err != nil {
		test.Fatal(err)
	}
	if string(fbuf) != "Test p" {
		test.Fatalf("invalid result parsed: '%s', expected: 'Test p'", string(fbuf))
	}

	if err := os.RemoveAll(tdir); err != nil {
		test.Error(err)
	}
}
