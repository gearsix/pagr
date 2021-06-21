package main

import (
	"notabug.org/gearsix/suti"
	"os"
	"path/filepath"
)

type Config struct {
	Contents  string
	Templates string
	Output    string
}

func (cfg *Config) relPaths(dir string) {
	var paths = []string{cfg.Contents, cfg.Templates, cfg.Output}
	for i, path := range paths {
		if !filepath.IsAbs(path) {
			paths[i] = filepath.Join(dir, path)
		}
	}
	cfg.Contents = paths[0]
	cfg.Templates = paths[1]
	cfg.Output = paths[2]
	return
}

// NewConfig returns a Config with default values
func NewConfig() Config {
	return Config{
		Contents:  "./content",
		Templates: "./templates",
		Output:    "./out",
	}
}

// NewConfigFromFile returns a Config with values read from the config file found at `fpath`.
// If values from the file are missing, default values are used.
// suti.LoadDataFile() is called to load the file (see notabug.org/gearsix/suti).
// Any relative filepaths in the returned Config are set relative to the parent directory of `fpath`.
func NewConfigFromFile(fpath string) (cfg Config, err error) {
	cfg = NewConfig()

	if _, err = os.Stat(fpath); err != nil {
		return
	}

	if err = suti.LoadDataFile(fpath, &cfg); err != nil {
		return
	}

	cfg.relPaths(filepath.Dir(fpath))
	return
}
