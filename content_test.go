package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadContentDir(test *testing.T) {
	test.Parallel()

	var err error
	tdir := filepath.Join(os.TempDir(), "pagr_test_TestLoadContentDir")
	if err := os.MkdirAll(tdir, 0775); err != nil {
		test.Errorf("failed to create temporary test dir: %s", tdir)
	}
	if err = createTestContents(tdir); err != nil {
		test.Errorf("failed to create test content: %s", err)
	}

	var p []Page
	if p, err = LoadContentDir(tdir); err != nil {
		test.Fatalf("LoadContentDir failed: %s", err)
	}

	validateTestPages(test, p, err)

	if err = os.RemoveAll(tdir); err != nil {
		test.Error(err)
	}
}

func TestNewContentFromFile(test *testing.T) {
	test.Parallel()

	var err error
	contents := map[string]string{
		"txt":  `test`,
		"md":   "**test**\ntest",
		"gfm":  "**test**\ntest",
		"cm":   "**test**",
		"html": `<b>test</b>`,
	}

	tdir := filepath.Join(os.TempDir(), "pagr_test", "TestNewContentFromFile")
	if err := os.MkdirAll(tdir, 0775); err != nil {
		test.Errorf("failed to create temporary test dir: %s", tdir)
	}
	contentsPath := func(ftype string) string {
		return tdir + "/test." + ftype
	}

	for ftype, data := range contents {
		if err = ioutil.WriteFile(contentsPath(ftype), []byte(data), 0666); err != nil {
			test.Error("TestNewContentFromFile setup failed:", err)
		}
	}

	var p Page
	for ftype := range contents {
		if err = p.NewContentFromFile(contentsPath(ftype)); err != nil {
			test.Fatal("NewContentFromFile failed for", ftype, err)
		}
	}

	if err = os.RemoveAll(tdir); err != nil {
		test.Error(err)
	}
}
