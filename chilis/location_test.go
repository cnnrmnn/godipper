package chilis

import (
	"errors"
	"testing"

	"github.com/antchfx/htmlquery"
)

var nearestTests = []struct {
	path string
	id   string
}{
	{"testdata/location1.html", "001.005.0945"},
	{"testdata/location2.html", "001.005.1094"},
	{"testdata/location3.html", "001.005.1320"},
}

var noneTests = []string{
	"testdata/location_none1.html",
	"testdata/location_none2.html",
}

var noDeliveryTests = []string{
	"testdata/location_no_delivery1.html",
	"testdata/location_no_delivery2.html",
}

func TestParseNearestID(t *testing.T) {
	for _, test := range nearestTests {
		path := test.path
		doc, err := htmlquery.LoadDoc(path)
		if err != nil {
			t.Errorf("%s: %v", path, err)
		}
		id, err := parseNearestID(doc)
		if err != nil {
			t.Errorf("%s: %v", path, err)
		}
		if id != test.id {
			t.Errorf("%s: id = %s, want %s", path, id, test.id)
		}
	}
}

func TestParseNearestIDNone(t *testing.T) {
	for _, test := range noneTests {
		reason := "no locations in proximity"
		doc, err := htmlquery.LoadDoc(test)
		if err != nil {
			t.Errorf("%s: %v", test, err)
		}
		_, err = parseNearestID(doc)
		if err == nil {
			t.Errorf("%s: err = nil, want (ForbiddenError) %s", test, reason)
		}
		var e ForbiddenError
		if !errors.As(err, &e) || reason != e.Reason {
			t.Errorf("%s: err = %v, want (ForbiddenError) %s", test, err, reason)
		}
	}
}

func TestParseNearestIDNoDelivery(t *testing.T) {
	for _, test := range noDeliveryTests {
		reason := "location doesn't deliver"
		doc, err := htmlquery.LoadDoc(test)
		if err != nil {
			t.Errorf("%s: %v", test, err)
		}
		_, err = parseNearestID(doc)
		if err == nil {
			t.Errorf("%s: err = nil, want (ForbiddenError) %s", test, reason)
		}
		var e ForbiddenError
		if !errors.As(err, &e) || reason != e.Reason {
			t.Errorf("%s: err = %s, want (ForbiddenError) %s", test, err, reason)
		}
	}
}

func TestParseNearestIDNoOrders(t *testing.T) {
	// No need to do more than one
	path := "testdata/location_no_orders.html"
	reason := "location is not accepting online orders"
	doc, err := htmlquery.LoadDoc(path)
	if err != nil {
		t.Errorf("%s: %v", path, err)
	}
	_, err = parseNearestID(doc)
	if err == nil {
		t.Errorf("%s: err = nil, want (ForbiddenError) %s", path, reason)
	}
	var e ForbiddenError
	if !errors.As(err, &e) || reason != e.Reason {
		t.Errorf("%s: err = %v, want (ForbiddenError) %s", path, err, reason)
	}
}

func TestParseLocation(t *testing.T) {
	// No need to do more than one
	path := "testdata/confirmation.html"
	test := Location{
		Name:  "Durham 15/501",
		Phone: "(919) 489-6699",
		Address: Address{
			Street: "4600 Chapel Hill Blvd.",
			City:   "Durham",
			State:  "NC",
			Zip:    "27707",
		},
	}
	doc, err := htmlquery.LoadDoc(path)
	if err != nil {
		t.Errorf("%s: %v", path, err)
	}
	location, err := parseLocation(doc)
	if err != nil {
		t.Errorf("%s: %v", path, err)
	}
	if location != test {
		t.Errorf("%s: location = %s, want %s", path, location, test)
	}
}
