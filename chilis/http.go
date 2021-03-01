package chilis

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"golang.org/x/net/publicsuffix"
)

// This should never return an error.
var chilisUrl, _ = url.Parse("https://www.chilis.com")

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
	jar.SetCookies(chilisUrl, []*http.Cookie{session})
	return jar, nil
}

// CreateClient creates and returns a client with a jar with the given session
// cookie in it.
func createClient(session *http.Cookie) (*http.Client, error) {
	jar, err := createSessionJar(session)
	if err != nil {
		return nil, fmt.Errorf("creating client: %v", err)
	}
	return &http.Client{Jar: jar}, err
}

// sessionID finds and returns the value of the session cookie given an HTTP
// client.
func sessionID(client *http.Client) (string, error) {
	for _, cookie := range client.Jar.Cookies(chilisUrl) {
		if cookie.Name == "SESSION" {
			return cookie.Value, nil
		}
	}
	return "", errors.New("failed to find session cookie")
}
