package main

import (
	"strings"

	"golang.org/x/net/html"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sitemap", func() {

	// getTokenizer(url string) (io.Reader, error)
	Describe("getTokenizer", func() {

		Context("when url is unreachable", func() {
			It("should return a nil repsonse and non-nil error", func() {
				resp, err := getTokenizer("http://doesnotexist.no")
				Ω(resp).Should(BeNil())
				Ω(err).ShouldNot(BeNil())
			})
		})
		Context("when url is reachable", func() {
			It("should return a non-nil response and a nil error", func() {
				resp, err := getTokenizer("http://news.ycombinator.com")
				Ω(resp).ShouldNot(BeNil())
				Ω(err).Should(BeNil())
			})
		})
	})

	// isValidToken(token html.TokenType) bool
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

	// isCorrectTag(token html.Token, tag string) bool
	Describe("isCorrectTag", func() {
		Context("when passed a token and tag that match", func() {
			It("should return true", func() {
				tag := "a"
				mock := `<a href="foo">link</a>`
				tokens := html.NewTokenizer(strings.NewReader(mock))
				tokens.Next()
				token := tokens.Token()
				Ω(isCorrectTag(token, tag)).Should(BeTrue())
			})
		})
		Context("when passed a token and tag that don't match", func() {
			It("should return false", func() {
				tag := "a"
				mock := `<br/>`
				tokens := html.NewTokenizer(strings.NewReader(mock))
				tokens.Next()
				token := tokens.Token()
				Ω(isCorrectTag(token, tag)).Should(BeFalse())
			})
		})
	})

	// getAttribute(token html.Token, attr string) (string, errror)
	Describe("getAttribute", func() {

		BeforeEach(func() {
			mock := `<a href="foo">link</a>`
			tokens := html.NewTokenizer(strings.NewReader(mock))
			tokens.Next()
			token := tokens.Token()
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
})
