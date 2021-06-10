package main

import (
    "fmt"
    "os"
    "testing"
)

func TestNewConfigFromFile(test *testing.T) {
    test.Parallel()
    tdir := test.TempDir()
    cfgp := fmt.Sprintf("%s/%s.toml", tdir, Name)
    if f, err := os.Create(cfgp); err != nil {
        test.Skipf("failed to create config file: '%s'", cfgp)
    } else {
        f.WriteString(`Output = "./test"`)
        f.Close()
    }

    if cfg, err := NewConfigFromFile(cfgp); err != nil {
        test.Fatal(err)
    } else if cfg.Output != tdir+"/test" {
        test.Fatalf(".Output invalid: '%s'", cfg.Output)
    }
}
