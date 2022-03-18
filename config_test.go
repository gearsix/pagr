package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// func TestNewConfig(test *testing.T) {} // don't waste time

func TestNewConfigFromFile(test *testing.T) {
	test.Parallel()

	tdir := filepath.Join(os.TempDir(), "pagr_test", "TestNewConfigFromFile")
	if err := os.MkdirAll(tdir, 0775); err != nil {
		test.Errorf("failed to create temporary test dir: %s", tdir)
	}
	cfgp := fmt.Sprintf("%s/%s.toml", tdir, Name)

	if f, err := os.Create(cfgp); err != nil {
		test.Skipf("failed to create config file: '%s'", cfgp)
	} else {
		f.WriteString(`
			Contents = "./p"
			Templates = "./t"
			Assets = ["./a"]
			Output = "./o"`)
		f.Close()
	}

	if cfg, err := NewConfigFromFile(cfgp); err == nil {
		if cfg.Contents != filepath.Join(tdir, "p") {
			test.Fatalf("invalid Contents path: '%s'", cfg.Contents)
		}
		if cfg.Templates != filepath.Join(tdir, "t") {
			test.Fatalf("invalid Templates path: '%s'", cfg.Templates)
		}
		if cfg.Assets[0] != filepath.Join(tdir, "a") {
			test.Fatalf("invalid Assets path: '%s'", cfg.Assets)
		}
		if cfg.Output != filepath.Join(tdir, "o") {
			test.Fatalf("invalid Output path: '%s'", cfg.Output)
		}
	} else {
		test.Fatal(err)
	}
	
	if err := os.RemoveAll(tdir); err != nil {
		test.Error(err)
	}
}
