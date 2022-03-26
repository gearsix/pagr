package main

import (
	"flag"
	"log"
	"notabug.org/gearsix/suti"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const Name = "pagr"
const Version = "0.0.0"

var gitBin string
var config Config
var flagConfig string
var flagVerbose bool

var ilog = log.New(os.Stdout, "", 0)
var elog = log.New(os.Stderr, "", 0)

func vlog(fmt string, args ...interface{}) {
	if flagVerbose {
		ilog.Printf(fmt, args...)
	}
}

func check(err error) {
	if err != nil {
		if flagVerbose {
			elog.Panic(err.Error())
		} else {
			elog.Fatalf("ERROR! %s\n", err)
		}
	}
}

func ignoreFile(filepath string) bool {
	return strings.Contains(filepath, ".ignore")
}

func init() {
	flag.BoolVar(&flagVerbose, "v", false, "print verbose ilog.")
	flag.StringVar(&flagConfig, "cfg", "", "path to pagr project configuration file")
	gitBin, _ = exec.LookPath("git")
}

func main() {
	flag.Parse()
	vlog("verbose on")
	config = loadConfigFile()
	vlog("loaded config: %s\n", config)

	var err error
	var content []Page
	content, err = LoadContentDir(config.Contents)
	check(err)
	ilog.Printf("loaded %d content pages", len(content))

	var templates []suti.Template
	templates, err = LoadTemplateDir(config.Templates)
	check(err)
	ilog.Printf("loaded %d template files", len(templates))

	ilog.Println("copying assets...")
	assetc := copyAssets()

	ilog.Println("building project...")
	pagec := 0
	for _, p := range content {
		vlog("+ %s", p.Path)
		
		_, err = p.Build(config.Output, findPageTemplate(p, templates))
		if err != nil {
			ilog.Printf("skipping %s: %s\n", p.Path, err)
			continue
		}

		for _, asset := range p.Assets {
			src := filepath.Join(config.Contents, asset)
			dst := filepath.Join(config.Output, asset)
			check(CopyFile(src, dst))
			vlog("\t-> %s\n", asset)
		}

		pagec++
		assetc += len(p.Assets)
	}

	ilog.Printf("generated %d html files, copied %d asset files\n", pagec, assetc)
	ilog.Println("pagr success")
	return
}

func loadConfigFile() Config {
	if len(flagConfig) > 0 {
		vlog("loading '%s'", flagConfig)
		c, err := NewConfigFromFile(flagConfig)
		check(err)
		return c
	} else {
		ilog.Println("no cfg passed, using defaults")
		return NewConfig()
	}
}

func findPageTemplate(p Page, t []suti.Template) (tmpl suti.Template) {
	for i, template := range t {
		if template.Name == p.TemplateName() {
			tmpl = t[i]
			break
		}
	}
	return
}

func copyAssets() (count int) {
	for _, asset := range config.Assets {
		asset = filepath.Clean(asset)
		filepath.Walk(asset,
			func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() && !ignoreFile(path) {
					dst := strings.TrimPrefix(filepath.Clean(path), asset)
					err = CopyFile(path, filepath.Join(config.Output, dst))
					vlog("\t-> %s\n", dst)
					count++
				}

				if err != nil {
					ilog.Printf("copy failed for %s: %s\n", path, err)
					err = nil
				}

				return err
			})
	}
	return
}


