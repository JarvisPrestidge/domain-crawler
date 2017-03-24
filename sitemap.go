package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

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
	startURL, err := getArgs()
	if err != nil {
		panic(err)
	}
	pageUrls, imageUrls, err := scrapeURL(startURL)
	if err != nil {
		panic(err)
	}
	for _, url := range pageUrls {
		fmt.Println(url)
	}
	fmt.Println("============================")
	for _, url := range imageUrls {
		fmt.Println(url)
	}
}

func scrapeURL(baseURL string) (pages, images []string, err error) {
	resp, err := http.Get(baseURL)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	content := html.NewTokenizer(resp.Body)
	for tt := content.Next(); isValidToken(tt); {
		if isStartTagToken(tt) {
			token := content.Token()
			page, err := scrapeToken(token, a, baseURL)
			if err == nil {
				pages = append(pages, page)
				goto next
			}
			image, err := scrapeToken(token, img, baseURL)
			if err == nil {
				pages = append(pages, image)
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

func fixURLProtocol(link string, baseURL string) (string, error) {
	switch {
	case link == "/":
		return baseURL, nil
	case link == "#":
		return "", errors.New("Redundant fragment found")
	case strings.Contains(link, "mailto:"):
		return "", errors.New("Mail scheme URI")
	case strings.Contains(link, "javascript:void(0)"):
		return "", errors.New("False link")
	case strings.Index(link, "http") == 0:
		return link, nil
	default:
		u, err := url.Parse(baseURL)
		if err != nil {
			return "", errors.New("Invalid URI")
		}
		// Constructing the base URL without paths / fragments
		home := u.Scheme + "://" + u.Host + "/"
		for link[:1] == "/" {
			if len(link) < 2 {
				break
			}
			link = strings.TrimPrefix(link, "/")
		}
		return home + link, nil
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
