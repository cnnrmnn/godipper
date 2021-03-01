package chilis

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"golang.org/x/net/publicsuffix"
)

// createJar creates and returns a cookie jar with secure options (public
// suffix list set)
func createJar() (*cookiejar.Jar, error) {
	options := &cookiejar.Options{PublicSuffixList: publicsuffix.List}
	jar, err := cookiejar.New(options)
	if err != nil {
		return nil, fmt.Errorf("creating cookie jar: %v")
	}
	return jar, nil
}

// createJar creates and returns a Jar with the given session cookie in it.
func createSessionJar(session *http.Cookie) (*cookiejar.Jar, error) {
	jar, err := createJar()
	if err != nil {
		return nil, fmt.Errorf("creating session cookie jar: %v")
	}
	u, err := url.Parse("https://www.chilis.com")
	if err != nil {
		return nil, fmt.Errorf("creating session cookie jar: %v", err)
	}
	jar.SetCookies(u, []*http.Cookie{session})
	return jar, nil
}

// createClient creates and returns a client with a jar with the given session
// cookie in it.
func createClient(session *http.Cookie) (*http.Client, error) {
	jar, err := createJar(session)
	if err != nil {
		return nil, fmt.Errorf("creating client: %v", err)
	}
	return &http.Client{Jar: jar}, err
}
