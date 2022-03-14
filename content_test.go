package main

import (
	"testing"
	"os"
)

func TestLoadContentsDir(test *testing.T) {
	var err error
	tdir := test.TempDir()
	if err = createTestContents(tdir); err != nil {
		test.Errorf("failed to create test content: %s", err)
	}

	var p []Page
	if p, err = LoadContentsDir(tdir); err != nil {
		test.Fatalf("LoadContentsDir failed: %s", err)
	}

	validateTestPages(test, p, err)
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
