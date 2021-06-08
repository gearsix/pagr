package main

import (
    "path/filepath"
    "io/fs"
    "strings"
    "notabug.org/gearsix/suti"
)

type Content []Page

func LoadContentDir(dirpath string) (c Content, e error) {
    pages := make(map[string]Page)
    defaults := make(map[string]Meta)
    e = filepath.Walk(dirpath, func(fpath string, info fs.FileInfo, err error) error {
        if err != nil {
            return nil
        }

        if info.IsDir() {
            p := NewPage(strings.TrimPrefix(fpath, dirpath))
            for _, dir := range strings.Split(fpath, "/") {
                if _, ok := defaults[dir]; ok {
                    p.Meta.MergeMeta(defaults[dir])
                }
            }
            return nil
        }

        pdir := filepath.Dir(fpath)
        page := pages[pdir]
        if strings.Contains(fpath, ".page") || strings.Contains(fpath, ".defaults") {
            var m Meta
            if err = suti.LoadDataFile(fpath, &m); err != nil {
                return err
            }
            if strings.Contains(fpath, ".page") {
                page.Meta.MergeMeta(m)
            } else if strings.Contains(fpath, ".defaults") {
                defaults[pdir] = m
            }
        } else if ext := filepath.Ext(fpath); ext == ".txt" || ext == ".md" || ext == ".html" {
            page.NewBodyFromFile(fpath)
        } else {
            page.Assets = append(page.Assets, strings.TrimPrefix(fpath, dirpath))
        }

        pages[pdir] = page
        return nil
    })

    for _, page := range pages {
        c = append(c, page)
    }

    return c, e
}

type Meta map[string]interface{}

func (m *Meta) MergeMeta(meta Meta, overwrite bool) {
    for k, v := range meta {
        if _, ok := m[k]; ok && overwrite {
            m[k] = v
        } else if !ok {
            m[k] = v
        }
    }
}

type Page struct {
    Path string
    Meta Meta
    Body []string
    Assets []string
}

func NewPage(path string) Page {
    return Page {
        Path: path,
        Meta: make(Meta),
        Body: make([]string, 0),
        Assets: make([]string, 0),
    }
}

func (p *Page) NewBodyFromFile(fpath string) {
}

