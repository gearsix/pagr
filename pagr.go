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

func main() {
	cfg := flag.String("cfg", "", "path to pagr project configuration file")
	flag.BoolVar(&verbose, "v", false, "print verbose logs")
	flag.Parse()

	vlog("verbose on")

	var err error

	var config Config
	if len(*cfg) > 0 {
		config, err = NewConfigFromFile(*cfg)
		check(err)
	} else {
		log.Println("warning: no cfg passed, using defaults")
		config = NewConfig()
	}

	var c Content
	c, err = LoadContentDir(config.Contents)
	check(err)

	var t []suti.Template
	t, err = LoadTemplateDir(config.Templates)
	check(err)

	build(config, c, t)

	return
}

func build(config Config, pages Content, templates []suti.Template) {
	var err error
	var out bytes.Buffer

	for _, pg := range pages {
		out.Reset()
		target := pg.GetTemplate()
		for _, t := range templates {
			tname := filepath.Base(t.Source)
			tname = strings.TrimSuffix(tname, filepath.Ext(tname))
			if tname == target {
				if out, err = t.Execute(pg); err != nil {
					log.Printf("Execution error in template '%s':\n", target)
					check(err)
				}
				outp := filepath.Join(config.Output, pg.Path, "index.html")
				check(os.MkdirAll(filepath.Dir(outp), 0755))
				check(os.WriteFile(outp, out.Bytes(), 0644))
				vlog("wrote '%s' -> '%s'", pg.Path, outp)
			}
		}
		if out.Len() == 0 {
			log.Printf("failed to find template '%s' for '%s'\n", target, pg.Path)
		}
	}
}
