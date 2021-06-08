package main

import (
    "flag"
    "io/fs"
    "log"
)

const Name = "pagr"
const Version = "0.0.0"

func main() {
    flag.StringVar(&cfg, "cfg", "", "path to pagr project configuration file")
    flag.BoolVar(&verbose, "verbose", false, "print verbose logs")
    flag.Parse()

    config, err := loadConfig(cfg)
    check(err)

    pages, err := loadContent(config.Contents)
    check(err)

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

func loadContent(root string) (p map[string]Page, e error) {
    p = make(map[string]Page)
    defaults := make(map[string]Meta)
    e = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
        if err != nil {
            return
        }

        if info.IsDir() {
            p := NewPage(strings.TrimPrefix(path, root))
            for _, dir := range strings.Split(path, "/") {
                if _, ok := defaults[dir]; ok {
                    p.MergeMeta(defaults[dir].Meta)
                }
            }
            return
        }

        pdir := filepath.Dir(path)
        page := p[pdir]

        if strings.Contains(path, ".page") || strings.Contains(path, ".defaults") {
            var m Meta
            err = suti.LoadDataFile(path, &m); err != nil {
                return err
            }
            if strings.Contains(path, ".page") {
                page.MergeMeta(m)
            } else if strings.Contains(path, ".defaults") {
                defaults[pdir] = m
            }
        } else if ext := filepath.Ext(path); ext == ".txt" || ext == ".md" || ext == ".html" {
            page.Body = append(page.Body, NewContentFromFile(path))
        } else {
            page.Assets = append(page.Assets, filepath.TrimPrefix(path, root))
        }

        p[pdir] = page
        return
    }
}

