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
	Name          string `json:"name"`
	StreetAddress string `json:"streetAddress"`
	City          string `json:"city"`
	State         string `json:"state"`
	Zip           string `json:"zip"`
	Phone         string `json:"phone"`
}

// NearestLocationID returns the ID of the nearest location that is in proximity
// of the given coordinates.
func NearestLocationID(lat, lng string) (string, error) {
	// Go documentation suggests that Clients should be reused rather than
	// created as needed due to internal state in their Transports. Address
	// this later. Client would ideally be reused for requests with same
	// session cookie, however those requests are triggered by HTTP
	// requests that could come at any time.
	client := &http.Client{}

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

// parseID parses and returns the location's ID from its root node if the
// location offers delivery.
func parseID(node *html.Node) (string, bool) {
	_, err := findOne(node, "span", "delivery icon-doordash")
	if err != nil {
		return "", false
	}

	id := htmlquery.SelectAttr(node, "id")[9:]
	if id == "" {
		return id, false
	}
	return id, true
}

// parseNearestID parses and returns the nearest location's ID, if any, from the
// location search page's root node.
func parseNearestID(doc *html.Node) (string, bool) {
	results, err := findOne(doc, "div", "col12 location-results")
	if err == nil {
		if nearest := results.FirstChild; nearest != nil {
			return parseID(nearest)
		}
	}
	return "", false
}
