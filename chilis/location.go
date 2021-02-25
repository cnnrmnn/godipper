package chilis

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// A Location is a Chili's restuarant location
type Location struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	StreetAddress string `json:"streetAddress"`
	Locality      string `json:"locality"`
	Region        string `json:"region"`
	PostalCode    string `json:"postalCode"`
	Delivery      bool   `json:"delivery"`
}

type Locations []Location

// Encode writes the JSON encoding of location to writer. At some point, write
// benchmark to see if this would run faster if written as a method of *Location
// (Location contains quite a bit of data).
func (location Location) Encode(writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	err := encoder.Encode(location)
	if err != nil {
		err = fmt.Errorf("writing JSON encoding of location: %v", err)
		return err
	}
	return err
}

// Encode writes the JSON encoding of locations to writer.
func (locations Locations) Encode(writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	err := encoder.Encode(locations)
	if err != nil {
		err = fmt.Errorf("writing JSON encoding of locations: %v", err)
		return err
	}
	return err
}

// FindLocations returns a slice of locations that are in proximity of the
// given coordinates.
func FindLocations(lat, lng string) (locations Locations) {
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

// spanInnerText finds the span tag with the given class and returns its inner
// text.
func spanInnerText(node *html.Node, class string) string {
	query := fmt.Sprintf("//span[@class='%v']", class)
	tag := htmlquery.FindOne(node, query)
	innerText := htmlquery.InnerText(tag)
	return innerText
}

// parseLocation parses and returns a location from the location's root node. At
// some point, write benchmark to see if this would run faster if written as a
// method of *Location (Location contains quite a bit of data).
func parseLocation(node *html.Node) Location {
	id := htmlquery.SelectAttr(node, "id")[9:]
	name := spanInnerText(node, "location-title")
	streetAddress := spanInnerText(node, "street-address")
	locality := spanInnerText(node, "locality")
	region := spanInnerText(node, "region")
	postalCode := spanInnerText(node, "postal-code")
	deliveryTag := htmlquery.FindOne(node, "//span[@class='delivery icon-doordash']")
	delivery := false
	if deliveryTag != nil {
		delivery = true
	}

	location := Location{
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
func parseLocations(doc *html.Node) (locations Locations) {
	results := htmlquery.FindOne(doc, "//div[@class=\"col12 location-results\"]")
	if results == nil {
		return locations
	}
	for result := results.FirstChild; result != nil; result = result.NextSibling {
		locations = append(locations, parseLocation(result))
	}
	return locations
}
