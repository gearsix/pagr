package main

import (
	"testing"
)

func TestBuildCrumbs(test *testing.T) {
}

func TestBuildSitemap(test *testing.T) {
	var err error

	tdir := test.TempDir()
	if err = createTestContents(tdir); err != nil {
		test.Errorf("failed to create test content: %s", err)
	}

	var p []Page
	if p, err = LoadContentsDir(tdir); err != nil {
		test.Errorf("LoadPagesDir failed: %s", err)
	}

	p = BuildSitemap(p)
	// TODO validate p
}
