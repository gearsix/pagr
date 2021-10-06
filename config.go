package main

import (
	"notabug.org/gearsix/suti"
	"os"
	"path/filepath"
)

// Config is the data structure containing all key/values to be loaded
// in pagr configuration files
type Config struct {
	Pages     string
	Templates string
	Assets    []string
	Output    string
}

// relPaths sets all filepath values in `cfg` relative to `dir`
func (cfg *Config) relPaths(dir string) {
	var paths = []string{cfg.Pages, cfg.Templates, cfg.Output}
	paths = append(paths, cfg.Assets...)
	for i, path := range paths {
		if !filepath.IsAbs(path) {
			paths[i] = filepath.Join(dir, path)
		}
	}
	cfg.Pages = paths[0]
	cfg.Templates = paths[1]
	cfg.Output = paths[2]
	cfg.Assets = paths[3:]
	return
}

// NewConfig returns a Config with default values
func NewConfig() Config {
	return Config{
		Pages:     "./content",
		Templates: "./templates",
		Assets:    []string{"./assets"},
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

	if err = suti.LoadDataFilepath(fpath, &cfg); err != nil {
		return
	}

	cfg.relPaths(filepath.Dir(fpath))
	return
}
