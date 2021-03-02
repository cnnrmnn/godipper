package chilis

import (
	"errors"
	"fmt"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// extras is a hash table that maps an Extra to its display name
var extras = map[Extra]string{
	0: "Ancho-Chile Ranch Dressing",
	1: "Avocado-Ranch Dressing",
	2: "Bleu Cheese Dressing",
	3: "Honey-Mustard Dressing",
	4: "Original BBQ Sauce",
	5: "Ranch Dressing",
}

// An Extra is an optional component of a Dipper, typically used for dipping
// sauces.
type Extra byte

// Name returns the Extra's display name.
func (e Extra) Name() string {
	return extras[e]
}

// parseID parses and returns an Extra's Chili's ID given its Item's Chili's ID.
func (e Extra) parseID(node *html.Node, iid string) (string, error) {
	var eid string
	// Groups of extras for the given item ID
	grps, err := find(node, attrQuery("div", "data-related", iid))
	if err != nil {
		return eid, fmt.Errorf("parsing Extra's Chili's ID: %v", err)
	}
	for _, grp := range grps {
		opt, err := findOne(grp, textQuery("option", e.Name()))
		if err == nil {
			eid = htmlquery.SelectAttr(opt, "value")
			return eid, nil
		}
	}
	return eid, errors.New("parsing Extra's Chili's ID")
}
