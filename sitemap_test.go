package main

import (
	"testing"
)

func TestBuildCrumbs(test *testing.T) {
}

func TestBuildSitemap(test *testing.T) {
	var err error
	/*
		writef := func(path, data string) {
			if err == nil {
				err = os.WriteFile(path, []byte(data), 0644)
			}
		}
	*/
	tdir := test.TempDir()
	if err = createProjectContents(tdir); err != nil {
		test.Errorf("failed to create test content: %s", err)
	}

	var p []Page
	if p, err = LoadPagesDir(tdir); err != nil {
		test.Errorf("LoadPagesDir failed: %s", err)
	}
	p = BuildSitemap(p)
	// TODO validate p
}
