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

func vlog(fmt string, args ...interface{}) {
	if verbose {
		log.Printf(fmt, args...)
	}
}

func check(err error) {
	if err != nil {
		if verbose {
			log.Panic(err.Error())
		} else {
			log.Fatalf("ERROR! %s\n", err)
		}
	}
}

func init() {
	flag.StringVar(&cfg, "cfg", "", "path to pagr project configuration file")
	flag.BoolVar(&verbose, "v", false, "print verbose logs")
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
		log.Println("warning: no cfg passed, using defaults")
		config = NewConfig()
	}
	vlog("loaded config: %s\n", config)

	var pages []Page
	pages, err = LoadPagesDir(config.Pages)
	check(err)
	log.Printf("loaded %d content pages", len(pages))

	var templates []suti.Template
	templates, err = LoadTemplateDir(config.Templates)
	check(err)
	log.Printf("loaded %d template files", len(templates))

	htmlc := 0
	var wg sync.WaitGroup
	assetc := copyAssets(wg, config)
	for _, page := range pages {
		wg.Add(1)
		go func (p Page) {
			defer wg.Done()
			if err := buildPage(config, p, templates); err != nil {
				log.Printf("skipping %s: %s\n", p.Path, err)
				return
			}
			check(p.CopyAssets(config.Pages, config.Output))
			vlog("-> %s", p.Path)
			htmlc++
			assetc += len(p.Assets)
		}(page)
	}
	wg.Wait()
	log.Printf("generated %d html files, copied %d asset files\n", htmlc, assetc)

	log.Println("pagr success")
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
		err := filepath.Walk(a, func(path string, info fs.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				n++
				wg.Add(1)
				go CopyFile(path, filepath.Join(cfg.Output, strings.TrimPrefix(path, filepath.Clean(a))))
			}
			return err
		})
		if !os.IsNotExist(err) {
			check(err)
		}
	}
	return n
}
