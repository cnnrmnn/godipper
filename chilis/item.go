package chilis

import (
	"fmt"

	"golang.org/x/net/html"
)

// An Item is a component of a triple dipper.
type Item interface {
	String() string
	ExtraValues() []string
}

// parseID parses and returns an Item's Chili's ID given its selection index.
func parseItemID(node *html.Node, val string, i int) (string, error) {
	var id string
	text := fmt.Sprintf("Selection %d", i+1)
	label, err := findOne(node, textQuery("label", text))
	if err != nil {
		return id, fmt.Errorf("parsing Item's Chili's ID: %v", err)
	}
	id, err = selectAttr(label.Parent, textQuery("option", val), "value")
	if err != nil {
		return id, fmt.Errorf("parsing Item's Chili's ID: %v", err)
	}
	return id, nil
}
