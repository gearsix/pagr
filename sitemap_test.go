package main

import (
	"testing"
)

func TestBuildCrumbs(test *testing.T) {
	var err error

	tdir := test.TempDir()
	if err = createTestContents(tdir); err != nil {
		test.Errorf("failed to create test content: %s", err)
	}

	// TODO fix laziness below, just dups TestLoadContentsDir
	var p []Page
	if p, err = LoadContentsDir(tdir); err != nil {
		test.Errorf("LoadContentsDir failed: %s", err)
	}
	
	validateTestPagesNav(test, p)
}

func TestBuildSitemap(test *testing.T) {
	var err error

	tdir := test.TempDir()
	if err = createTestContents(tdir); err != nil {
		test.Errorf("failed to create test content: %s", err)
	}

	// TODO fix laziness below, just dups TestLoadContentsDir
	var p []Page
	if p, err = LoadContentsDir(tdir); err != nil {
		test.Errorf("LoadContentsDir failed: %s", err)
	}
	
	validateTestPagesNav(test, p)
}
