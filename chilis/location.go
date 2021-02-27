package chilis

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// A Location is a Chili's restuarant location
type Location struct {
	Name    string  `json:"name"`
	Address Address `json:"address"`
	Phone   string  `json:"phone"`
}

// NearestLocationID returns the ID of the nearest location that is in proximity
// of the given coordinates.
func NearestLocationID(lat, lng string) (string, error) {
	client := http.DefaultClient

	req, err := http.NewRequest("GET", "https://www.chilis.com/locations/results", nil)
	if err != nil {
		err = fmt.Errorf("creating locations request: %v", err)
		return "", err
	}

	query := url.Values{
		"lat": []string{lat},
		"lng": []string{lng},
	}
	req.URL.RawQuery = query.Encode()

	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("fetching location: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		err = fmt.Errorf("parsing locations html: %v", err)
		return "", err
	}
	id, ok := parseNearestID(doc)
	if !ok {
		err = errors.New("no locations in proximity")
	}
	return id, err
}

// SetLocation sets the Chili's location for a new session and returns the new
// session cookie.
func SetLocation(id string) (*http.Cookie, error) {
	session, err := startSession()
	if err != nil {
		return nil, err
	}
	client := http.DefaultClient

	req, err := http.NewRequest("GET", "https://www.chilis.com/order", nil)
	if err != nil {
		err = fmt.Errorf("setting location: %v", err)
		return nil, err
	}
	req.AddCookie(session)
	query := url.Values{
		"rid": []string{id},
	}
	req.URL.RawQuery = query.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("setting location: %v", err)
	}
	resp.Body.Close()

	return session, nil
}

// startSession starts a Chili's session and returns the new session cookie.
// Unfortunately, the first request (before the session cookie is set) can't set
// the session's location.
func startSession() (*http.Cookie, error) {
	resp, err := http.Get("https://www.chilis.com")
	if err != nil {
		return nil, fmt.Errorf("starting session: %v", err)
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "SESSION" {
			return cookie, nil
		}
	}
	return nil, errors.New("failed to find session cookie")
}

// parseNearestID parses and returns the nearest location's ID, if any, from the
// location search page's root node.
func parseNearestID(doc *html.Node) (string, bool) {
	nearest, err := findOne(doc, classQuery("div", "location"))
	if err != nil {
		return "", false
	}
	_, err = findOne(nearest, classQuery("span", "delivery icon-doordash"))
	if err != nil {
		return "", false
	}

	id := htmlquery.SelectAttr(nearest, "id")[9:]
	if id == "" {
		return id, false
	}
	return id, true
}

// parseLocation parses and returns a Location from an order confirmation page.
func parseLocation(doc *html.Node) (location Location, ok bool) {
	node, err := findOne(doc, classQuery("div", "location-address-wrapper"))
	if err != nil {
		return location, false
	}
	// Assume no errors after locating address wrapper.
	street, _ := innerText(node, classQuery("div", "location-address-street"))
	city, _ := innerText(node, classQuery("span", "location-address-city"))
	state, _ := innerText(node, classQuery("span", "location-address-state"))
	zip, _ := innerText(node, classQuery("span", "location-address-zip"))
	phone, _ := innerText(node, classQuery("a", "location-phone tel"))
	return Location{
		Address: Address{
			Street: street,
			City:   city,
			State:  state,
			Zip:    zip,
		},
		Phone: phone,
	}, true
}
