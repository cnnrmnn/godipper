package chilis

import (
	"errors"
	"fmt"
	"net/http"
)

// A Session is composed of a Chili's session ID and an HTTP client with the
// session cookie set.
type Session struct {
	ID     string
	Client *http.Client
}

// NewSession returns a pointer to a new Session given a session ID.
func NewSession(id string) (*Session, error) {
	var sess *Session
	cook := http.Cookie{Name: "Session", Value: id}
	clt, err := createClient(&cook)
	if err != nil {
		return sess, fmt.Errorf("creating session: %v", err)
	}

	return &Session{id, clt}, err
}

// StartSession returns a pointer to a new Session.
func StartSession() (*Session, error) {
	jar, err := createJar()
	if err != nil {
		return nil, fmt.Errorf("starting session: %v", err)
	}
	clt := &http.Client{Jar: jar}
	resp, err := clt.Get("https://www.chilis.com")
	if err != nil {
		return nil, fmt.Errorf("starting session: %v", err)
	}
	resp.Body.Close()

	var id string
	for _, cook := range resp.Cookies() {
		if cook.Name == "SESSION" {
			id = cook.Value
		}
	}
	if id == "" {
		return nil, errors.New("failed to find session cookie")
	}
	return &Session{id, clt}, err
}
