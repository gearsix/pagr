package main

import (
	"flag"
	"log"
)

const Name = "pagr"
const Version = "0.0.0"

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

	var c Content
	c, err = LoadContentDir(config.Contents)
	check(err)

	log.Println(c)

	return
}

func check(err error) {
	if err != nil {
		log.Fatalf("ERROR! %s\n", err)
	}
}
