package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTemplateDir(t *testing.T) {
	t.Parallel()

	tdir := filepath.Join(os.TempDir(), "pagr_test", "TestLoadTemplateDir")
	if err := os.MkdirAll(tdir, 0775); err != nil {
		t.Errorf("failed to create temporary test dir: %s", tdir)
	}

	if err := createTestTemplates(tdir); err != nil {
		t.Errorf("failed to create test templates: %s", err)
	}

	tmpls, err := LoadTemplateDir(tdir)
	if err != nil {
		t.Fatal(err)
	}
	if len(tmpls) != len(templates)*2 { // * 2 for partials
		t.Fatalf("number of returned templates is %d (should be %d)",
			len(tmpls), len(templates))
	}

	if err = os.RemoveAll(tdir); err != nil {
		t.Error(err)
	}
}
