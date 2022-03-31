package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildCrumbs(test *testing.T) {
	var err error

	tdir := filepath.Join(os.TempDir(), "pagr_test", "TestBuildCrumbs")
	if err := os.MkdirAll(tdir, 0775); err != nil {
		test.Errorf("failed to create temporary test dir: %s", tdir)
	}
	if err = createTestContents(tdir); err != nil {
		test.Errorf("failed to create test content: %s", err)
	}

	// TODO fix laziness below, just dups TestLoadContentDir
	var p []Page
	if p, err = LoadContentDir(tdir); err != nil {
		test.Errorf("LoadContentDir failed: %s", err)
	}

	validateTestPagesNav(test, p)

	if err = os.RemoveAll(tdir); err != nil {
		test.Error(err)
	}
}

func TestBuildSitemap(test *testing.T) {
	var err error

	tdir := filepath.Join(os.TempDir(), "pagr_test", "TestBuildSitemap")
	if err := os.MkdirAll(tdir, 0775); err != nil {
		test.Errorf("failed to create temporary test dir: %s", tdir)
	}
	if err = createTestContents(tdir); err != nil {
		test.Errorf("failed to create test content: %s", err)
	}

	// TODO fix laziness below, just dups TestLoadContentDir
	var p []Page
	if p, err = LoadContentDir(tdir); err != nil {
		test.Errorf("LoadContentDir failed: %s", err)
	}

	validateTestPagesNav(test, p)

	if err = os.RemoveAll(tdir); err != nil {
		test.Error(err)
	}
}
