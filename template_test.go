package main

import (
	"fmt"
	"path/filepath"
	"os"
	"testing"
)

func TestLoadTemplateDir(t *testing.T) {
	t.Parallel()

	tdir := t.TempDir()

	var err error
	if err = createProjectTemplates(tdir); err != nil {
		t.Errorf("failed to create test templates: %s", err)
	}

	tmpls, err := LoadTemplateDir(tdir)
	if err != nil {
		t.Fatal(err)
	}
	if len(tmpls) != len(templates) * 2 { // * 2 for partials
		t.Fatalf("number of returned templates is %d (should be %d)",
			len(tmpls), len(templates))
	}
}

var templates = map[string]string{ // [ext]template
	"tmpl": "{{.Contents}}",
	"mst": "{{Contents}}",
}

func createProjectTemplates(dir string) error {
	var err error
	writef := func(path, data string) {
		if err == nil {
			err = os.WriteFile(path, []byte(data), 0644)
		}
	}

	for ext, data := range templates {
		writef(fmt.Sprintf("%s/%s.%s", dir, ext, ext), data)
		writef(fmt.Sprintf("%s/%s.ignore.%s", dir, ext, ext), data)
		writef(fmt.Sprintf("%s/%s.%s.ignore", dir, ext, ext), data)

		pdir := filepath.Join(dir, ext)
		err = os.Mkdir(pdir, 0755)
		writef(fmt.Sprintf("%s/partial.%s", pdir, ext), data)
		writef(fmt.Sprintf("%s/partial.ignore.%s", pdir, ext), data)
		writef(fmt.Sprintf("%s/partial.%s.ignore", pdir, ext), data)

		if err != nil {
			break
		}
	}

	return err
}
