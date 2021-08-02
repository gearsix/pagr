package main

import (
	"flag"
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
			log.Panic(err)
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

	var p []Page
	p, err = LoadPagesDir(config.Pages)
	check(err)
	log.Printf("loaded %d content pages", len(p))

	var t []suti.Template
	t, err = LoadTemplateDir(config.Templates)
	check(err)
	log.Printf("loaded %d template files", len(t))

	htmlc := 0
	assetc := 0
	var wg sync.WaitGroup
	for _, pg := range p {
		var tmpl suti.Template
		tmpl, err = findTemplate(pg, t)
		if os.IsNotExist(err) {
			log.Printf("warning: skipping '%s', failed to find template '%s'\n", pg.Path, pg.GetTemplate())
			continue
		} else {
			check(err)
		}
		wg.Add(1)
		go func(page Page) {
			defer wg.Done()
			check(page.Build(config.Output, tmpl))
			check(page.CopyAssets(config.Pages, config.Output))
			vlog("-> %s", page.Path)
		}(pg)
		htmlc++
		assetc += len(pg.Assets)
	}
	wg.Wait()
	log.Printf("generated %d html files, copied %d asset files\n", htmlc, assetc)

	log.Println("pagr success")
	return
}

func findTemplate(pg Page, templates []suti.Template) (suti.Template, error) {
	var t suti.Template
	err := os.ErrNotExist
	target := pg.GetTemplate()
	for _, t := range templates {
		tname := filepath.Base(t.Source)
		if tname == target || strings.TrimSuffix(tname, filepath.Ext(tname)) == target {
			return t, nil
		}
	}
	return t, err
}

