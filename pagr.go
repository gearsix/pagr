package main

import (
	"flag"
	"log"
	"notabug.org/gearsix/suti"
)

const Name = "pagr"
const Version = "0.0.0"

func check(err error) {
	if err != nil {
		log.Fatalf("ERROR! %s\n", err)
	}
}

func main() {
	cfg := flag.String("cfg", "", "path to pagr project configuration file")
	//verbose := flag.Bool("verbose", false, "print verbose logs")
	flag.Parse()

	var err error

	var config Config
	if len(*cfg) > 0 {
		config, err = NewConfigFromFile(*cfg)
		check(err)
	} else {
		log.Println("warning: no cfg passed, using defaults")
		config = NewConfig()
	}

	var _ Content
	_, err = LoadContentDir(config.Contents)
	check(err)

	var _ []suti.Template
	_, err = LoadTemplateDir(config.Templates)
	check(err)

	return
}
