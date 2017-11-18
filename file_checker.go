package main

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
	"gopkg.in/russross/blackfriday.v2"
)

type fileChecker struct {
	urlChecker urlChecker
}

func newFileChecker(timeout time.Duration, s semaphore) fileChecker {
	return fileChecker{newURLChecker(timeout, s)}
}

func (c fileChecker) Check(f string) ([]urlResult, error) {
	n, err := parseFile(f)

	if err != nil {
		return nil, err
	}

	us := extractURLs(n)
	rc := make(chan urlResult, len(us))
	rs := make([]urlResult, 0, len(us))

	go c.urlChecker.CheckMany(us, rc)

	for r := range rc {
		rs = append(rs, r)
	}

	return rs, nil
}

func (c fileChecker) CheckMany(fs []string, rc chan<- fileResult) {
	wg := sync.WaitGroup{}

	for _, f := range fs {
		wg.Add(1)

		go func(f string) {
			if rs, err := c.Check(f); err == nil {
				rc <- fileResult{filename: f, urlResults: rs}
			} else {
				rc <- fileResult{filename: f, err: err}
			}

			wg.Done()
		}(f)
	}

	wg.Wait()
	close(rc)
}

func parseFile(f string) (*html.Node, error) {
	bs, err := ioutil.ReadFile(f)

	if err != nil {
		return nil, err
	}

	if !isHTMLFile(f) {
		bs = blackfriday.Run(bs)
	}

	n, err := html.Parse(bytes.NewReader(bs))

	if err != nil {
		return nil, err
	}

	return n, nil
}

func extractURLs(n *html.Node) []string {
	us := make(map[string]bool)
	ns := make([]*html.Node, 0, 1024)
	ns = append(ns, n)

	for len(ns) > 0 {
		i := len(ns) - 1
		n := ns[i]
		ns = ns[:i]

		if n.Type == html.ElementNode {
			switch n.Data {
			case "a":
				for _, a := range n.Attr {
					if a.Key == "href" && isURL(a.Val) {
						us[a.Val] = true
						break
					}
				}
			case "img":
				for _, a := range n.Attr {
					if a.Key == "src" && isURL(a.Val) {
						us[a.Val] = true
						break
					}
				}
			}
		}

		for n := n.FirstChild; n != nil; n = n.NextSibling {
			ns = append(ns, n)
		}
	}

	return stringSetToSlice(us)
}

func isURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}

func isHTMLFile(f string) bool {
	return strings.HasSuffix(f, ".html") || strings.HasSuffix(f, ".htm")
}
