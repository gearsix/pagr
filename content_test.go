package main

import (
	"fmt"
	"os"
	"testing"
)

func TestLoadContentDir(t *testing.T) {
	t.Parallel()

	var err error
	tdir := t.TempDir()
	if err = createProjectContents(tdir); err != nil {
		t.Errorf("failed to create test content: %s", err)
	}

	var c Content
	if c, err = LoadContentDir(tdir); err != nil {
		t.Fatalf("LoadContentDir failed: %s", err)
	}

	validateContents(t, c, err)
}

func validateContents(t *testing.T, c Content, e error) {
	if len(c) != len(contents)-1 {
		t.Fatalf("invalid number of pages returned (%d should be %d)",
			len(c), len(contents))
	}

	for _, p := range c {
		if len(p.Path) == 0 {
			t.Fatalf("empty Path for page:\n%s\n", p)
		}
		if _, ok := p.Meta["test"]; !ok || len(p.Meta) == 0 {
			t.Fatalf("empty Meta for page:\n%s\n", p)
		}
		if len(p.Contents) == 0 {
			t.Fatalf("empty Contents for page:\n%s\n", p)
		}
		if len(p.Assets) == 0 {
			t.Fatalf("empty Assets for page:\n%s\n", p)
		}
	}
}

var contents = map[string]string{
	".txt": `p1
p2

	pre1
	pre2
p3

p4
`,
	".html": `<p>p1<br>
p2</p>
<pre>pre1
pre2
</pre>
<p>p3</p>
<p>p4</p>`,
	".md": `p1
p2

	pre1
	pre2

p3
`,
	".gfm": `p1
p2

	pre1
	pre2

p3`,
	".cm": `p1
p2

	pre1
	pre2

p3`,
}

var asset = []byte{ // 5x5 black png
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00,
	0x0d, 0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00,
	0x00, 0x05, 0x01, 0x03, 0x00, 0x00, 0x00, 0xb7, 0xa1, 0xb4, 0xa6,
	0x00, 0x00, 0x00, 0x06, 0x50, 0x4c, 0x54, 0x45, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0xa5, 0x67, 0xb9, 0xcf, 0x00, 0x00, 0x00, 0x02,
	0x74, 0x52, 0x4e, 0x53, 0xff, 0x00, 0xe5, 0xb7, 0x30, 0x4a, 0x00,
	0x00, 0x00, 0x09, 0x70, 0x48, 0x59, 0x73, 0x00, 0x00, 0x0e, 0xc4,
	0x00, 0x00, 0x0e, 0xc4, 0x01, 0x95, 0x2b, 0x0e, 0x1b, 0x00, 0x00,
	0x00, 0x10, 0x49, 0x44, 0x41, 0x54, 0x08, 0x99, 0x63, 0x60, 0x66,
	0x60, 0x66, 0x60, 0x00, 0x62, 0x76, 0x00, 0x00, 0x4a, 0x00, 0x11,
	0x3a, 0x34, 0x8c, 0xad, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
	0x44, 0xae, 0x42, 0x60, 0x82,
}

func createProjectContents(dir string) (err error) {
	writef := func(path, data string) {
		if err == nil {
			err = os.WriteFile(path, []byte(data), 0644)
		}
	}

	for l, lang := range SupportedContent {
		if l == 0 {
			writef(fmt.Sprintf("%s/.defaults.json", dir), "{ \"test\": \"data\" }")
			writef(fmt.Sprintf("%s/.page.toml", dir), "test = \"data\"")
		} else if l > 1 {
			dir, err = os.MkdirTemp(dir, "page")
		}

		writef(fmt.Sprintf("%s/body%d%s", dir, l, lang), contents[lang])
		writef(fmt.Sprintf("%s/asset.png", dir), string(asset))

		if err != nil {
			break
		}
	}

	return
}
