package main

import (
    "flag"
    "log"
)

const Name = "pagr"
const Version = "0.0.0"

var flagDir string
var flagDirDesc = "directory of target project"
var flagVerbose bool
var flagVerboseDesc = "print verbose logs"

func init() {
    flag.StringVar(&flagDir, "d", ".", flagDirDesc)
    flag.StringVar(&flagDir, "dir", ".", flagDirDesc)
    flag.BoolVar(&flagVerbose, "v", false, flagVerboseDesc)
    flag.BoolVar(&flagVerbose, "verbose", false, flagVerboseDesc)
    flag.Parse()
}

func main() {
}

func check(error) {
    if err != nil {
        log.Printf("ERROR! %s\n", err)
    }
    return err != nil
}
