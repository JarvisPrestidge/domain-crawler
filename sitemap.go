package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"

	"golang.org/x/net/html"
)

const (
	linkTag  = "a"
	imageTag = "img"
	hrefAttr = "href"
	srcAttr  = "src"
	ncpu     = 8
)

type node struct {
	tag, attr string
}

var a = node{tag: linkTag, attr: hrefAttr}
var img = node{tag: imageTag, attr: srcAttr}

func main() {
	runtime.GOMAXPROCS(ncpu)
	startURL, err := getArgs()
	if err != nil {
		panic(err)
	}
	c := make(chan string)
	go func() {
		for url := range c {
			go crawl(url, c)
		}
	}()
	c <- startURL
	var input string
	fmt.Scanln(&input)
	fmt.Println("Done")
}

func crawl(baseURL string, c chan<- string) {
	resp, err := http.Get(baseURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	content := html.NewTokenizer(resp.Body)
	for tt := content.Next(); isValidToken(tt); {
		if isStartTagToken(tt) {
			token := content.Token()
			page, err := scrapeToken(token, a, baseURL)
			if err == nil {
				if ok := withinDomain(page, baseURL); ok {
					fmt.Println(page)
					c <- page
				}
				goto next
			}
			image, err := scrapeToken(token, img, baseURL)
			if err == nil {
				fmt.Println(image)
				goto next
			}
		}
	next:
		tt = content.Next()
	}
	return
}

func scrapeToken(token html.Token, node node, baseURL string) (string, error) {
	if isTag(token, node.tag) {
		val, err := getAttribute(token, node.attr)
		if err != nil {
			return "", errors.New("No attribute value found")
		}
		url, err := fixURLProtocol(val, baseURL)
		if err != nil {
			return "", err
		}
		return url, nil
	}
	return "", errors.New("Un-wanted tag")
}

func withinDomain(link string, baseURL string) bool {
	domain, _ := url.Parse(baseURL)
	candidate, _ := url.Parse(link)
	return domain.Host == candidate.Host
}

func fixURLProtocol(link string, baseURL string) (string, error) {
	if ok := handleEdgeCases(link); !ok {
		return "", errors.New("False link edge case caught")
	}
	if strings.Index(link, "http") == 0 {
		return link, nil
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", errors.New("Invalid URI")
	}
	home := u.Scheme + "://" + u.Host + "/"
	for link[:1] == "/" {
		if len(link) < 2 {
			break
		}
		link = strings.TrimPrefix(link, "/")
	}
	return home + link, nil
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
