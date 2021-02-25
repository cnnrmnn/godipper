package chilis

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// A Location is a Chili's restuarant location
type Location struct {
	ID            string
	Name          string
	StreetAddress string
	Locality      string
	Region        string
	PostalCode    string
	Delivery      bool
}

// FindLocations returns a slice of locations that are in proximity of the
// given coordinates.
func FindLocations(lat, lng string) (locations []*Location) {
	// Go documentation suggests that Clients should be reused rather than
	// created as needed due to internal state in their Transports. Address
	// this later. Client would ideally be reused for requests with same
	// session cookie, however those requests are triggered by HTTP
	// requests that could come at any time.
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://www.chilis.com/locations/results", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating locations request: %v\n", err)
		return
	}

	query := url.Values{
		"lat": []string{lat},
		"lng": []string{lng},
	}
	req.URL.RawQuery = query.Encode()

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error fetching locations: %v\n", err)
		return
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing locations html: %v\n", err)
		return
	}
	return parseLocations(doc)
}

// parseLocation parses and returns a location from the location's root node.
func parseLocation(node *html.Node) (location *Location) {
	id := htmlquery.SelectAttr(node, "id")[9:]
	nameTag := htmlquery.FindOne(node, "//span[@class='location-title']")
	name := htmlquery.InnerText(nameTag)
	addressTag := htmlquery.FindOne(node, "//span[@class='street-address']")
	streetAddress := htmlquery.InnerText(addressTag)
	localityTag := htmlquery.FindOne(node, "//span[@class='locality']")
	locality := htmlquery.InnerText(localityTag)
	regionTag := htmlquery.FindOne(node, "//span[@class='region']")
	region := htmlquery.InnerText(regionTag)
	postalCodeTag := htmlquery.FindOne(node, "//span[@class='postal-code']")
	postalCode := htmlquery.InnerText(postalCodeTag)
	deliveryTag := htmlquery.FindOne(node, "//span[@class='delivery icon-doordash']")
	delivery := false
	if deliveryTag != nil {
		delivery = true
	}

	location = &Location{
		ID:            id,
		Name:          name,
		StreetAddress: streetAddress,
		Locality:      locality,
		Region:        region,
		PostalCode:    postalCode,
		Delivery:      delivery,
	}
	return location
}

// parseLocation parses and returns a slice of locations from the location
// search page's root node.
func parseLocations(doc *html.Node) (locations []*Location) {
	results := htmlquery.FindOne(doc, "//div[@class=\"col12 location-results\"]")
	if results == nil {
		return locations
	}
	for result := results.FirstChild; result != nil; result = result.NextSibling {
		locations = append(locations, parseLocation(result))
	}
	return locations
}
