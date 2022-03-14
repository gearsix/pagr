package main

import (
	"strings"
	"sort"
)

func findRootPage(pages []Page) (root *Page) {
	for i, p := range pages {
		if p.Path == "/" {
			root = &pages[i]
			break
		}
	}
	return
}

func readPageDepth(p Page) (depth int) {
	if p.Path == "/" {
		depth = 0
	} else {
		depth = len(strings.Split(p.Path, "/")[1:])
	}
	return
}

// BuildCrumbs will generate a `[]*Page`, where each item is a pointer to the Page
// found `pages`, who's `.Path` matches a crumb in `p.Path`.
// "crumbs" referring to https://en.wikipedia.org/wiki/Breadcrumb_navigation
func BuildCrumbs(p Page, pages []Page) (crumbs []*Page) {
	var path string
	for _, c := range strings.Split(p.Path, "/")[1:] {
		path += "/" + c
		for j, pp := range pages {
			if pp.Path == path {
				crumbs = append(p.Nav.Crumbs, &pages[j])
				break
			}
		}
	}
	return
}

// Sitemap parses `pages` to determine the `.Nav` values for each element in `pages`
// based on their `.Path` value. These values will be set in the returned Content
func BuildSitemap(pages []Page) []Page {
	root := findRootPage(pages)
	
	for i, p := range pages {
		pdepth := readPageDepth(p)
		
		p.Nav.Root = root
		
		if pdepth == 1 && p.Path != "/" {
			p.Nav.Parent = root
		}
		
		for j, pp := range pages {
			ppdepth := readPageDepth(pp)
		
			p.Nav.All = append(p.Nav.All, &pages[j])
			
			if p.Nav.Parent == nil && ppdepth == pdepth-1 && strings.Contains(p.Path, pp.Path) {
				p.Nav.Parent = &pages[j]
			}
			if ppdepth == pdepth+1 && strings.Contains(pp.Path, p.Path) {
				p.Nav.Children = append(p.Nav.Children, &pages[j])
			}
		}
		
		sort.SliceStable(p.Nav.Children, func(i, j int) bool {
			return sort.StringsAreSorted([]string{p.Nav.Children[j].Path, p.Nav.Children[j].Path})
		})
		
		p.Nav.Crumbs = BuildCrumbs(p, pages)
		
		pages[i] = p
	}
	
	return pages
}
