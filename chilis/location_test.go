package chilis

import (
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
