package chilis

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// A Location is a Chili's restuarant location
type Location struct {
	Name    string  `json:"name"`
	Phone   string  `json:"phone"`
	Address Address `json:"address"`
}

// SetLocation sets the Chili's location for a new session and returns an HTTP
// client for future requests.
func SetLocation(addr Address) (*http.Client, error) {
	clt, err := startSession()
	if err != nil {
		return nil, fmt.Errorf("setting location: %v", err)
	}
	id, err := nearestLocationID(clt, addr)
	if err != nil {
		return nil, fmt.Errorf("setting location: %v", err)
	}
	u := fmt.Sprintf("https://www.chilis.com/order?rid=%s", id)

	resp, err := clt.Get(u)
	if err != nil {
		return nil, fmt.Errorf("setting location: %v", err)
	}
	resp.Body.Close()

	return clt, nil
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
	fmt.Println(u.String())

	resp, err := clt.Get(u.String())
	if err != nil {
		return id, fmt.Errorf("fetching location: %v", err)
	}
	defer resp.Body.Close()
	//	yeh, _ := io.ReadAll(resp.Body)
	//	fmt.Println(string(yeh))

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return id, fmt.Errorf("parsing locations html: %v", err)
	}
	id, ok := parseNearestID(doc)
	if !ok {
		return id, ForbiddenError{"no locations in proximity"}
	}
	return id, err
}

// startSession starts a Chili's session and returns an HTTP client for future
// requests. Unfortunately, the first request (before the session cookie is
// set) can't set the session's location.
func startSession() (*http.Client, error) {
	jar, err := createJar()
	if err != nil {
		return nil, fmt.Errorf("starting session: %v", err)
	}
	client := &http.Client{Jar: jar}
	_, err = client.Get("https://www.chilis.com")
	if err != nil {
		return nil, fmt.Errorf("starting session: %v", err)
	}
	return client, nil
}

// parseNearestID parses and returns the nearest location's ID, if any, from the
// location search page's root node.
func parseNearestID(doc *html.Node) (string, bool) {
	var id string
	nearest, err := findOne(doc, classQuery("div", "location"))
	if err != nil {
		return id, false
	}
	_, err = findOne(nearest, classQuery("span", "delivery icon-doordash"))
	if err != nil {
		return id, false
	}

	id = htmlquery.SelectAttr(nearest, "id")[9:]
	if id == "" {
		return id, false
	}
	return id, true
}

// parseLocation parses and returns a Location from an order confirmation page.
func parseLocation(doc *html.Node) (Location, error) {
	var loc Location
	wrp, err := findOne(doc, classQuery("div", "location-address-wrapper"))
	if err != nil {
		return loc, fmt.Errorf("parsing location: %v")
	}
	// Don't use functions that return errors for nil results after wrapper is
	// located.
	fo := htmlquery.FindOne
	it := htmlquery.InnerText
	name := it(fo(wrp, classQuery("div", "location-name")))
	street := it(fo(wrp, classQuery("div", "location-address-street")))
	city := it(fo(wrp, classQuery("span", "location-address-city")))
	state := it(fo(wrp, classQuery("span", "location-address-state")))
	zip := it(fo(wrp, classQuery("span", "location-address-zip")))
	phone := it(fo(wrp, classQuery("a", "location-phone tel")))
	return Location{
		Name:  name,
		Phone: phone,
		Address: Address{
			Street: street,
			City:   city,
			State:  state,
			Zip:    zip,
		},
	}, nil
}
