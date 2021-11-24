package main

import (
	"fmt"
	"flag"
	"io/fs"
	"os"
	"strings"
	"sync"
	"path/filepath"
	"log"
	"notabug.org/gearsix/suti"
)

const Name = "pagr"
const Version = "0.0.0"

var cfg string
var verbose bool
var ilog = log.New(os.Stdout, "", 0)
var elog = log.New(os.Stderr, "", 0)

func vlog (fmt string, args ...interface{}) {
	if verbose {
		ilog.Printf(fmt, args...)
	}
}

func check(err error) {
	if err != nil {
		if verbose {
			elog.Panic(err.Error())
		} else {
			elog.Fatalf("ERROR! %s\n", err)
		}
	}
}

func ignoreFile(filepath string) bool {
	return strings.Contains(filepath, ".ignore")
}

func init() {
	flag.StringVar(&cfg, "cfg", "", "path to pagr project configuration file")
	flag.BoolVar(&verbose, "v", false, "print verbose ilog.")
}

func main() {
	flag.Parse()
	vlog("verbose on")

	var err error
	var config Config
	if len(cfg) > 0 {
		vlog("loading '%s'", cfg)
		config, err = NewConfigFromFile(cfg)
		check(err)
	} else {
		ilog.Println("no cfg passed, using defaults")
		config = NewConfig()
	}
	vlog("loaded config: %s\n", config)

	var pages []Page
	pages, err = LoadPagesDir(config.Pages)
	check(err)
	ilog.Printf("loaded %d content pages", len(pages))

	var templates []suti.Template
	templates, err = LoadTemplateDir(config.Templates)
	check(err)
	ilog.Printf("loaded %d template files", len(templates))

	ilog.Println("building project...")
	htmlc := 0
	var wg sync.WaitGroup
	assetc := copyAssets(wg, config)
	for _, page := range pages {
		if err := buildPage(config, page, templates); err != nil {
			ilog.Printf("skipping %s: %s\n", page.Path, err)
			return
		}
		check(page.CopyAssets(config.Pages, config.Output))
		vlog("-> %s", page.Path)
		htmlc++
		assetc += len(page.Assets)
	}
	ilog.Printf("generated %d html files, copied %d asset files\n", htmlc, assetc)

	ilog.Println("pagr success")
	return
}

func findTemplateIndex(p Page, templates []suti.Template) (t int) {
	for t, template := range templates {
		if template.Name == p.TemplateName() {
			return t
		}
	}
	return -1
}

func buildPage(cfg Config, p Page, t []suti.Template) error {
	var tmpl *suti.Template
	for i, template := range t {
		if template.Name == p.TemplateName() {
			tmpl = &t[i]
		}
	}
	if tmpl == nil {
		return fmt.Errorf("failed to find template '%s'", p.TemplateName())
	}

	_, err := p.Build(cfg.Output, *tmpl)
	check(err)
	check(p.CopyAssets(cfg.Pages, cfg.Output))
	return err
}

func copyAssets(wg sync.WaitGroup, cfg Config) (n int) {
	for _, a := range cfg.Assets {
		err := filepath.Walk(a, func(src string, info fs.FileInfo, err error) error {
			if err == nil && !info.IsDir() && !ignoreFile(src) {
				a = filepath.Clean(a)
				path := strings.TrimPrefix(src, a)
				n++
				check(CopyFile(src, filepath.Join(cfg.Output, path)))
				vlog("-> %s", path)
			}
			return err
		})
		if !os.IsNotExist(err) {
			check(err)
		}
	}
	return n
}
