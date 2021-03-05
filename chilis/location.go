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
	Phone   string  `json:"phone"`
	Address Address `json:"address"`
}

// SetLocation sets the Chili's location for a new session and returns a
// Session..
func SetLocation(addr Address) (*Session, error) {
	sess, err := StartSession()
	if err != nil {
		return nil, fmt.Errorf("setting location: %v", err)
	}
	clt := sess.Client
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

	return sess, nil
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

// parseNearestID parses and returns the nearest location's ID, if any, from the
// location search page's root node.
func parseNearestID(doc *html.Node) (string, error) {
	var id string
	nearest, err := findOne(doc, classQuery("div", "location"))
	if err != nil {
		return id, ForbiddenError{"no locations in proximity"}
	}
	// XPath query
	q := "//a[@class='btn slim order-btn' and text()='Order Now']"
	_, err = findOne(nearest, q)
	if err != nil {
		return id, ForbiddenError{"location is not accepting online orders"}
	}
	_, err = findOne(nearest, classQuery("span", "delivery icon-doordash"))
	if err != nil {
		return id, ForbiddenError{"location doesn't deliver"}
	}

	id = htmlquery.SelectAttr(nearest, "id")[9:]
	if id == "" {
		return id, errors.New("parsing nearest location ID")
	}
	return id, nil
}

// parseLocation parses and returns a Location from an order confirmation page.
func parseLocation(doc *html.Node) (Location, error) {
	var loc Location
	wrp, err := findOne(doc, classQuery("div", "location-address-wrapper"))
	if err != nil {
		return loc, fmt.Errorf("parsing location: %v", err)
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
