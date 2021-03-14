package chilis

import (
	"errors"
	"fmt"

	"golang.org/x/net/html"
)

// parseID parses and returns an Extra's Chili's ID given its Item's Chili's ID.
func parseExtraID(node *html.Node, val, iid string) (string, error) {
	var eid string
	// Groups of extras for the given item ID
	grps, err := find(node, attrQuery("div", "data-related", iid))
	if err != nil {
		return eid, fmt.Errorf("parsing Extra's Chili's ID: %v", err)
	}
	for _, grp := range grps {
		eid, err := selectAttr(grp, textQuery("option", val), "value")
		if err != nil {
			continue
		}
		return eid, nil
	}
	return eid, errors.New("parsing Extra's Chili's ID")
}
