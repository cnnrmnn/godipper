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

// classQuery returns an XPath query for an HTML element with the given name and
// class attribute.
func classQuery(name string, class string) string {
	return fmt.Sprintf("//%s[@class='%s']", name, class)
}

// innerText finds the first HTML element with the given name and class
// attribute and returns its inner text.
func innerText(node *html.Node, name string, class string) (string, error) {
	query := classQuery(name, class)
	tag := htmlquery.FindOne(node, query)
	if tag == nil {
		return "", errors.New("no matching elements found")
	}
	innerText := htmlquery.InnerText(tag)
	return innerText
}

// parseID parses and returns the location's ID from its root node if the
// location offers delivery.
func parseID(node *html.Node) (string, bool) {
	delivery := htmlquery.FindOne(node, "//span[@class='delivery icon-doordash']")
	if delivery == nil {
		return "", false
	}

	return htmlquery.SelectAttr(node, "id")[9:], true
}

// parseNearestID parses and returns the nearest location's ID, if any, from the
// location search page's root node.
func parseNearestID(doc *html.Node) (string, bool) {
	results := htmlquery.FindOne(doc, "//div[@class=\"col12 location-results\"]")
	if results != nil {
		if nearest := results.FirstChild; nearest != nil {
			return parseID(nearest)
		}
	}
	return "", false
}
