package chilis

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
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

// SetLocation sets the Chili's location for session.
func (s *Session) SetLocation(addr Address) error {
	clt := s.Client
	id, err := nearestLocationID(clt, addr)
	if err != nil {
		return fmt.Errorf("setting location: %v", err)
	}
	u := fmt.Sprintf("https://www.chilis.com/order?rid=%s", id)

	resp, err := clt.Get(u)
	if err != nil {
		return fmt.Errorf("setting location: %v", err)
	}
	resp.Body.Close()

	return nil
}

// nearestLocationID returns the ID of the nearest location that is in proximity
// of the given address.
func nearestLocationID(clt *http.Client, addr Address) (string, error) {
	var id string

	u, err := url.Parse("https://www.chilis.com/locations/results")
	if err != nil {
		return id, fmt.Errorf("parsing location URL: %v", err)
	}
	query := url.Values{
		"query": []string{addr.String()},
	}
	u.RawQuery = query.Encode()

	resp, err := clt.Get(u.String())
	if err != nil {
		return id, fmt.Errorf("fetching location: %v", err)
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return id, fmt.Errorf("parsing locations html: %v", err)
	}
	return parseNearestID(doc)
}
