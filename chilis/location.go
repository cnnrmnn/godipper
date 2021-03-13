package chilis

import (
	"errors"
	"fmt"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// A Location is a Chili's restuarant location
type Location struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
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

	name, err := innerText(wrp, classQuery("div", "location-name"))
	if err != nil {
		return loc, fmt.Errorf("parsing location name: %v", err)
	}
	phone, err := innerText(wrp, classQuery("a", "location-phone tel"))
	if err != nil {
		return loc, fmt.Errorf("parsing location phone: %v", err)
	}
	return Location{
		Name:  name,
		Phone: phone,
	}, nil
}
