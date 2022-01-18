package main

import (
	"fmt"
	"path/filepath"
	"os"
	"testing"
)

// func TestNewConfig(test *testing.T) {} // don't waste time

func TestNewConfigFromFile(test *testing.T) {
	test.Parallel()

	tdir := test.TempDir()
	cfgp := fmt.Sprintf("%s/%s.toml", tdir, Name)

	if f, err := os.Create(cfgp); err != nil {
		test.Skipf("failed to create config file: '%s'", cfgp)
	} else {
		f.WriteString(`
			Pages = "./p"
			Templates = "./t"
			Assets = ["./a"]
			Output = "./o"`)
		f.Close()
	}

	if cfg, err := NewConfigFromFile(cfgp); err == nil {
		if cfg.Pages != filepath.Join(tdir, "p") {
			test.Fatalf("invalid Pages path: '%s'", cfg.Pages)
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
}
