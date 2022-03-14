package main

import (
	"testing"
)

func TestLoadTemplateDir(t *testing.T) {
	t.Parallel()

	tdir := t.TempDir()

	var err error
	if err = createTestTemplates(tdir); err != nil {
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
}
