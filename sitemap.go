package main

import (
	"bufio"
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

// structs representing <a> and <img> elements
var a = node{tag: linkTag, attr: hrefAttr}
var img = node{tag: imageTag, attr: srcAttr}

func main() {
	// Get command line args
	arg, err := getArgs()
	check(err)
	// Clean incoming url
	startURL, domain, err := formatURL(arg)
	fmt.Println(startURL)
	check(err)
	// Declare maps and channel for urls
	pages := make(map[string]struct{})
	images := make(map[string]struct{})
	cp := make(chan string)
	ci := make(chan string)
	// Setup output file / dir
	f := setupFile(domain)
	defer f.Close()
	w := bufio.NewWriter(f)
	// Setup waitgroup for gorountines
	var wg sync.WaitGroup
	wg.Add(1)
	counter := 0
	// Closure gorourtine to listen on incoming channels
	go func() {
		for {
			// listen on channel with select until timeout
			select {
			case page := <-cp:
				// check if already scraped
				if _, ok := pages[page]; !ok {
					// if not, add it to pages, crawl it and write to output
					pages[page] = struct{}{}
					wg.Add(1)
					go crawl(page, cp, ci, &wg)
					fmt.Fprintln(w, page)
					counter++
				}
			case <-time.After(time.Second * 5):
				// After 5 seconds of no incoming messages finish
				wg.Done()
				break
			}
		}
	}()
	// same for the image channel without crawling
	go func() {
		for image := range ci {
			if _, ok := images[image]; !ok {
				images[image] = struct{}{}
				fmt.Fprintln(w, image)
				counter++
			}
		}
	}()
	// Pass out starting url onto the page channel
	cp <- startURL
	// Wait for goroutines to finish
	wg.Wait()
	// Flush urls to file
	w.Flush()
	fmt.Println("Successfully crawled", counter, "links!")
}

// crawl input url and report back on seperate channels the valid crawlalbe uls
func crawl(baseURL string, cp, ci chan<- string, wg *sync.WaitGroup) {
	// decrement the waitgroup counter on function exit
	defer wg.Done()
	// get the response
	resp, err := http.Get(baseURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	content := html.NewTokenizer(resp.Body)
	// iterate over valid tokens
	for tt := content.Next(); isValidToken(tt); {
		// look only for start tokens i.e. <a> or <img>
		if isStartTagToken(tt) {
			// scrpae the attirbutes from the tags
			token := content.Token()
			page, err := scrapeToken(token, a, baseURL)
			if err == nil {
				// check with the domain
				if ok := isWithinDomain(page, baseURL); ok {
					cp <- page
				}
				goto next
			}
			// same for images
			image, err := scrapeToken(token, img, baseURL)
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

// scrapeToken takes a token and a node and returns the corresponding url attirbute
func scrapeToken(token html.Token, node node, baseURL string) (string, error) {
	// check its the html tag we want
	if isTag(token, node.tag) {
		// then get the attribute
		val, err := getAttribute(token, node.attr)
		if err != nil {
			return "", errors.New("No attribute value found")
		}
		// fix the protocol for none absolute and false urls
		url, err := formatURLProtocol(val, baseURL)
		if err != nil {
			return "", err
		}
		return url, nil
	}
	return "", errors.New("Un-wanted tag")
}

// isWithinDomain checks if two urls are from a common host domain
func isWithinDomain(link string, baseURL string) bool {
	domain, err := url.Parse(baseURL)
	if err != nil {
		return false
	}
	candidate, err := url.Parse(link)
	if err != nil {
		return false
	}
	return domain.Host == candidate.Host
}

// formatURLProtocol fixes none absolute urls and handles url related edge cases
func formatURLProtocol(link string, baseURL string) (string, error) {
	switch {
	case !handleEdgeCases(link):
		return "", errors.New("False link edge case caught")
	case strings.Index(link, "http") == 0:
		return link, nil
	case strings.Index(link, "www") == 0:
		return "http://" + link, nil
	default:
		// below is for internal (i.e. none absolute) url
		u, err := url.Parse(link)
		if err != nil {
			return "", errors.New("Invalid URI")
		}
		base, err := url.Parse(baseURL)
		if err != nil {
			return "", errors.New("Invalid URI")
		}
		// resolve and merge the url paths
		return base.ResolveReference(u).String(), nil
	}
}

// handleEdgeCases covers the cases of backward facing urls
// fragments, invalid schemes and javascript rendered links
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

// getAttribute returns the string values of the corresponding token attirbutes
func getAttribute(token html.Token, attr string) (string, error) {
	// iterate over all the attributes
	for _, value := range token.Attr {
		// until we find the one we're looking for and break
		if value.Key == attr {
			if len(value.Val) >= 1 {
				return value.Val, nil
			}
			break
		}
	}
	return "", errors.New("No attribute with key: " + attr + " in token")
}

// getArgs get the input url to start the crawler
func getArgs() (string, error) {
	if args := len(os.Args); args == 2 {
		return os.Args[1], nil
	}
	return "", errors.New("Incorrect command line arguments passed")
}

// formatURL fixes that input urls from getArgs() or errors
// it returnt the fixed link and the links host domain without paths
func formatURL(link string) (string, string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", "", errors.New("Invalid URL")
	}
	// if the link isn't absolute then prepend with a valid scheme
	if !u.IsAbs() {
		link = "http://" + link
	}
	d, err := url.Parse(link)
	if err != nil {
		return "", "", errors.New("Invalid URL")
	}
	return d.String(), d.Host, nil
}

func setupFile(domain string) *os.File {
	// Creating file name
	s := strings.Split(domain, ".")
	filename := s[len(s)-2]
	// Create dir if not exists
	if _, err := os.Stat("output"); os.IsNotExist(err) {
		os.Mkdir("output", os.ModePerm)
	}
	// Create an output file
	file, err := os.Create("output/" + filename + ".sitemap")
	check(err)
	return file
}

// isValidToken checks if a token type is an error token (i.e. eof)
func isValidToken(tt html.TokenType) bool {
	return tt != html.ErrorToken
}

//isStartTagToken checks if a token type is a start tag (i.e. <a>)
func isStartTagToken(tt html.TokenType) bool {
	return tt == html.StartTagToken
}

// isTag checks that a token is infact the html element we want based on param 'tag'
func isTag(token html.Token, tag string) bool {
	return token.Data == tag
}

// check is a helper error function
func check(e error) {
	if e != nil {
		panic(e)
	}
}
