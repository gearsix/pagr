package main

import (
    "fmt"
    "flag"
    "log"
)

const Name = "pagr"
const Version = "0.0.0"

func init() {
}

func main() {
    flag.StringVar(&cfg, "cfg", "", "path to pagr project configuration file")
    flag.BoolVar(&verbose, "verbose", false, "print verbose logs")
    flag.Parse()

    config, err := loadConfig(cfg)
    check(err)
    fmt.Println(config)

    return
}

func check(err error) {
    if err != nil {
        log.Fatalf("ERROR! %s\n", err)
    }
}

func loadConfig(fpath string) (c Config, e error) {
    if len(cfg) > 0 {
        c, e = NewConfigFromFile(cfg)
    } else {
        c = NewConfig()
    }
    return
}

