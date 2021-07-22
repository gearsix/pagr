package main

import (
	"bytes"
	"os"
	"strings"
	"flag"
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

	var c Content
	c, err = LoadContentDir(config.Contents)
	check(err)
	log.Printf("loaded %d content pages", len(c))

	var t []suti.Template
	t, err = LoadTemplateDir(config.Templates)
	check(err)
	log.Printf("loaded %d template files", len(t))

	build(config, c, t)
	log.Println("pagr success")

	return
}

func build(config Config, pages Content, templates []suti.Template) {
	var err error
	var out bytes.Buffer

	outc := len(pages)
	for _, pg := range pages {
		out.Reset()
		target := pg.GetTemplate()
		for _, t := range templates {
			tname := filepath.Base(t.Source)
			if tname == target || strings.TrimSuffix(tname, filepath.Ext(tname)) == target {
				if out, err = t.Execute(pg); err != nil {
					log.Printf("Execution error in template '%s':\n", target)
					check(err)
				}
				outp := filepath.Join(config.Output, pg.Path, "index.html")
				check(os.MkdirAll(filepath.Dir(outp), 0755))
				check(os.WriteFile(outp, out.Bytes(), 0644))
				vlog("-> %s", outp)
			}
		}
		if out.Len() == 0 {
			log.Printf("warning: skipping '%s', failed to find template '%s'\n", pg.Path, target)
			outc--
		}
	}
	log.Printf("generated %d html files\n", outc)
}
