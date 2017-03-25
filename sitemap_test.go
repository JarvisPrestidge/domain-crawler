package main

import (
	"strings"

	"golang.org/x/net/html"

	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sitemap", func() {

	Describe("isStartTagToken", func() {

		var tt html.TokenType
		var tokens *html.Tokenizer

		BeforeEach(func() {
			mock := `<a href="foo">link</a>`
			tokens = html.NewTokenizer(strings.NewReader(mock))
			tt = tokens.Next()
		})

		Context("when passed a start tag token type", func() {
			It("should return true", func() {
				Ω(isStartTagToken(tt)).Should(BeTrue())
			})
		})
		Context("when passed a non start tag toke type", func() {
			It("should return false", func() {
				tt = tokens.Next()
				Ω(isStartTagToken(tt)).Should(BeFalse())
			})
		})
	})

	Describe("isValidToken", func() {

		Context("when passed an valid token", func() {
			It("should return true", func() {
				mock := `<a href="foo">link</a>`
				tokens := html.NewTokenizer(strings.NewReader(mock))
				tt := tokens.Next()
				Ω(isValidToken(tt)).Should(BeTrue())
			})
		})
		Context("when passed an invalid token", func() {
			It("should return false", func() {
				mock := `<br/>`
				tokens := html.NewTokenizer(strings.NewReader(mock))
				tokens.Next()
				tt := tokens.Next()
				Ω(isValidToken(tt)).Should(BeFalse())
			})
		})
	})

	Describe("isTag", func() {
		Context("when passed a token and tag that match", func() {
			It("should return true", func() {
				tag := "a"
				mock := `<a href="foo">link</a>`
				tokens := html.NewTokenizer(strings.NewReader(mock))
				tokens.Next()
				token := tokens.Token()
				Ω(isTag(token, tag)).Should(BeTrue())
			})
		})
		Context("when passed a token and tag that don't match", func() {
			It("should return false", func() {
				tag := "a"
				mock := `<br/>`
				tokens := html.NewTokenizer(strings.NewReader(mock))
				tokens.Next()
				token := tokens.Token()
				Ω(isTag(token, tag)).Should(BeFalse())
			})
		})
	})

	Describe("getAttribute", func() {
		var token html.Token

		BeforeEach(func() {
			mock := `<a href="foo">link</a>`
			tokens := html.NewTokenizer(strings.NewReader(mock))
			tokens.Next()
			token = tokens.Token()
		})

		Context("when passed a token with a matching attribute", func() {
			It("should return the value of that attribute", func() {
				attr := "href"
				res, err := getAttribute(token, attr)
				Ω(res).Should(Equal("foo"))
				Ω(err).Should(BeNil())
			})
		})
		Context("when passed a token without a matching attribute", func() {
			It("should return an non-nil errror", func() {
				attr := "class"
				_, err := getAttribute(token, attr)
				Ω(err).ShouldNot(BeNil())
			})
		})
	})

	Describe("getArgs", func() {
		Context("when one argument has been passed", func() {
			It("should return the arg and a nil error message", func() {
				os.Args = append(os.Args, "test")
				val, err := getArgs()
				Ω(val).Should(Equal("test"))
				Ω(err).Should(BeNil())
			})
		})
		Context("when 2 arguments are passed", func() {
			It("should return a non-nil error message", func() {
				os.Args = append(os.Args, "another test")
				_, err := getArgs()
				Ω(err).ShouldNot(BeNil())
			})
		})
		Context("when 0 arguments are passed", func() {
			It("should return a non-nil error message", func() {
				os.Args = os.Args[:len(os.Args)-2]
				_, err := getArgs()
				Ω(err).ShouldNot(BeNil())
			})
		})
	})

	//func formatURL(link string) (string, string, error) {
	Describe("formatURL", func() {
		Context("when url without scheme is passed", func() {
			It("should return the fixed url with valid http scheme, the domain without paths and a nil error message", func() {
				// no http scheme
				url, domain, err := formatURL("golangweekly.com/issue1")
				Ω(url).Should(Equal("http://golangweekly.com/issue1"))
				Ω(domain).Should(Equal("golangweekly.com"))
				Ω(err).Should(BeNil())
			})
		})
		Context("when url without scheme is passed", func() {
			It("should return the same url, the domain without paths and a nil eror message", func() {
				url, domain, err := formatURL("http://golangweekly.com")
				Ω(url).Should(Equal("http://golangweekly.com"))
				Ω(domain).Should(Equal("golangweekly.com"))
				Ω(err).Should(BeNil())
			})
		})
		Context("when passed an invalid url", func() {
			It("should return a non-nil error message", func() {
				_, _, err := formatURL("somet\"hing::/sj")
				Ω(err).ShouldNot(BeNil())
			})
		})
	})

	// func handleEdgeCases(link string) bool {
	Describe("handleEdgeCases", func() {
		Context("when passed a invalid url strings", func() {
			It("should return false", func() {
				Ω(handleEdgeCases("/")).Should(BeFalse())
				Ω(handleEdgeCases("#")).Should(BeFalse())
				Ω(handleEdgeCases("mailto:jarvisprestidge@gmail.com")).Should(BeFalse())
				Ω(handleEdgeCases("https://something.com/javascript:void(0)")).Should(BeFalse())
			})
		})
		Context("when passed any other url", func() {
			It("should return true", func() {
				Ω(handleEdgeCases("http://golangweekly.com")).Should(BeTrue())
				Ω(handleEdgeCases("golangweekly.com")).Should(BeTrue())
				Ω(handleEdgeCases("www.golangweekly.com")).Should(BeTrue())
			})
		})
	})

	// func formatURLProtocol(link string, baseURL string) (string, error) {
	Describe("formatURLProtocol", func() {

		mock := "https://golangweekly.com/issues1"

		Context("when passed a broken url", func() {
			It("should return a non-nil error message", func() {
				_, err := formatURLProtocol("/", "http://golangweekly.com")
				Ω(err).ShouldNot(BeNil())
				_, err1 := formatURLProtocol("#", "http://golangweekly.com")
				Ω(err1).ShouldNot(BeNil())
				_, err2 := formatURLProtocol("mailto:jarvisprestidge@gmail.com", "http://golangweekly.com")
				Ω(err2).ShouldNot(BeNil())
				_, err3 := formatURLProtocol("https://golangweekly.com/javascript:void(0)", "http://golangweekly.com")
				Ω(err3).ShouldNot(BeNil())
			})
		})
		Context("when passed a url with a valid http scheme", func() {
			It("should return the same link passed in", func() {
				url, err := formatURLProtocol(mock, "http://golangweekly.com")
				Ω(url).Should(Equal(mock))
				Ω(err).Should(BeNil())
			})
		})
		Context("when passed a url without a valid scheme, but www.", func() {
			It("should return that link prefixed with a valid http scheme", func() {
				url, err := formatURLProtocol("www.golangweekly.com", "http://golangweekly.com")
				Ω(url).Should(Equal("http://www.golangweekly.com"))
				Ω(err).Should(BeNil())
			})
		})
		Context("when passed a non absolute url", func() {
			It("should resolve the nonabs path with the base url and return it", func() {
				url, err := formatURLProtocol("/this-is-an/non-abs-path", "http://golangweekly.com")
				Ω(url).Should(Equal("http://golangweekly.com/this-is-an/non-abs-path"))
				Ω(err).Should(BeNil())
			})
		})
	})

	// func isWithinDomain(link string, baseURL string) bool {
	Describe("isWithinDomain", func() {
		Context("when passed 2 urls with the same host domain", func() {
			It("should return true", func() {
				ok := isWithinDomain("http://golangweekly.com/issues1", "http://golangweekly.com")
				Ω(ok).Should(BeTrue())
			})
		})
		Context("when passed 2 urls with different host domains", func() {
			It("should return false", func() {
				ok := isWithinDomain("http://new.ycombinator.com/news", "http://golangweekly.com")
				Ω(ok).Should(BeFalse())
			})
		})
	})

	// func scrapeToken(token html.Token, node node, baseURL string) (string, error) {
	Describe("scrapeToken", func() {
		Context("when passed a link token and node that match", func() {
			It("should return the attribute url and a nil error message", func() {
				mock := `<a href="http://golangweekly.com/issue1">link</a>`
				tokens := html.NewTokenizer(strings.NewReader(mock))
				tokens.Next()
				token := tokens.Token()
				url, err := scrapeToken(token, node{tag: "a", attr: "href"}, "http://golangweekly.com")
				Ω(url).Should(Equal("http://golangweekly.com/issue1"))
				Ω(err).Should(BeNil())
			})
		})
		Context("when passed an img token and node that match", func() {
			It("should return the attribute url and a nil error message", func() {
				mock := `<img src="http://golangweekly.com/something.png">image</img>`
				tokens := html.NewTokenizer(strings.NewReader(mock))
				tokens.Next()
				token := tokens.Token()
				url, err := scrapeToken(token, node{tag: "img", attr: "src"}, "http://golangweekly.com")
				Ω(url).Should(Equal("http://golangweekly.com/something.png"))
				Ω(err).Should(BeNil())
			})
		})
		Context("when passed a token and node that match but without an valid attribute", func() {
			It("should return a non-nil error message", func() {
				mock := `<img>image</img>`
				tokens := html.NewTokenizer(strings.NewReader(mock))
				tokens.Next()
				token := tokens.Token()
				_, err := scrapeToken(token, node{tag: "img", attr: "src"}, "http://golangweekly.com")
				Ω(err).ShouldNot(BeNil())
			})
		})
	})
})
