package chilis

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"golang.org/x/net/publicsuffix"
)

// createJar creates and returns a Jar with the given session cookie in it.
func createJar(session *http.Cookie) (*cookiejar.Jar, error) {
	options := &cookiejar.Options{PublicSuffixList: publicsuffix.List}
	jar, err := cookiejar.New(options)
	if err != nil {
		return nil, fmt.Errorf("adding TripleDipper to cart: %v")
	}
	u, err := url.Parse("https://www.chilis.com")
	if err != nil {
		return nil, fmt.Errorf("creating CookieJar: %v", err)
	}
	jar.SetCookies(u, []*http.Cookie{session})
	return jar, nil
}
