package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

const (
	linkTag  = "a"
	imageTag = "img"
	hrefAttr = "href"
	srcAttr  = "src"
)

type node struct {
	tag, attr string
}

var a = node{tag: linkTag, attr: hrefAttr}
var img = node{tag: imageTag, attr: srcAttr}

func main() {
	// Get command line args
	startURL, err := getArgs()
	check(err)
	startURL, err = fixURLProtocol(startURL, startURL)
	check(err)
	// Maps for holding references to urls
	pages := make(map[string]struct{})
	images := make(map[string]struct{})
	// Channel for pages and images
	cp := make(chan string)
	ci := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			select {
			case page := <-cp:
				if _, ok := pages[page]; !ok {
					pages[page] = struct{}{}
					wg.Add(1)
					go crawl(page, cp, ci, &wg)
					fmt.Println(page)
				}
			case <-time.After(time.Second * 5):
				wg.Done()
				break
			}
		}
	}()
	go func() {
		for image := range ci {
			if _, ok := images[image]; !ok {
				images[image] = struct{}{}
				fmt.Println(image)
			}
		}
	}()
	cp <- startURL
	wg.Wait()
	fmt.Println("Done!")
}

func crawl(startURL string, cp, ci chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := http.Get(startURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	content := html.NewTokenizer(resp.Body)
	for tt := content.Next(); isValidToken(tt); {
		if isStartTagToken(tt) {
			token := content.Token()
			page, err := scrapeToken(token, a, startURL)
			if err == nil {
				if ok := withinDomain(page, startURL); ok {
					cp <- page
				}
				goto next
			}
			image, err := scrapeToken(token, img, startURL)
			if err == nil {
				ci <- image
				goto next
			}
		}
	next:
		tt = content.Next()
	}
	return
}

func scrapeToken(token html.Token, node node, startURL string) (string, error) {
	if isTag(token, node.tag) {
		val, err := getAttribute(token, node.attr)
		if err != nil {
			return "", errors.New("No attribute value found")
		}
		url, err := fixURLProtocol(val, startURL)
		if err != nil {
			return "", err
		}
		return url, nil
	}
	return "", errors.New("Un-wanted tag")
}

func withinDomain(link string, startURL string) bool {
	domain, err := url.Parse(startURL)
	if err != nil {
		return false
	}
	candidate, err := url.Parse(link)
	if err != nil {
		return false
	}
	return domain.Host == candidate.Host
}

func fixURLProtocol(link string, startURL string) (string, error) {
	switch {
	case !handleEdgeCases(link):
		return "", errors.New("False link edge case caught")
	case strings.Index(link, "http") == 0:
		return link, nil
	case strings.Index(link, "www") == 0:
		return "http://" + link, nil
	default:
		d, err := url.Parse(startURL)
		if err != nil {
			return "", errors.New("Invalid URI")
		}
		home := d.Scheme + "://" + d.Host + "/"
		for link[:1] == "/" {
			if len(link) < 2 {
				break
			}
			link = strings.TrimPrefix(link, "/")
		}
		return home + link, nil
	}
}

func handleEdgeCases(link string) bool {
	switch {
	case link == "/":
		return false
	case link == "#":
		return false
	case strings.Contains(link, "mailto:"):
		return false
	case strings.Contains(link, "javascript:"):
		return false
	default:
		return true
	}
}

func getAttribute(token html.Token, attr string) (string, error) {
	for _, value := range token.Attr {
		if value.Key == attr {
			if len(value.Val) >= 1 {
				return value.Val, nil
			}
			break
		}
	}
	return "", errors.New("No attribute with key: " + attr + " in token")
}

func getArgs() (string, error) {
	if args := len(os.Args); args == 2 {
		return os.Args[1], nil
	}
	return "", errors.New("Incorrect command line arguments passed")
}

func isValidToken(tt html.TokenType) bool {
	return tt != html.ErrorToken
}

func isStartTagToken(tt html.TokenType) bool {
	return tt == html.StartTagToken
}

func isTag(token html.Token, tag string) bool {
	return token.Data == tag
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
