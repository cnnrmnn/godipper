package chilis

import (
	"errors"
	"fmt"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// items is a hash table that maps an Item to its display name.
var items = map[Item]string{
	0:  "Awesome Blossom Petals",
	1:  "Big Mouth® Bites",
	2:  "Boneless Buffalo Wings",
	3:  "Boneless Honey-Chipotle Wings",
	4:  "Boneless House BBQ Wings",
	5:  "Boneless Mango-Habanero Wings",
	6:  "Buffalo Wings",
	7:  "Crispy Cheddar Bites",
	8:  "Crispy Chicken Crispers",
	9:  "Crispy Honey-Chipotle Chicken Crispers®",
	10: "Crispy Mango-Habanero Crispers®",
	11: "Fried Pickles",
	12: "Honey-Chipotle Wings",
	13: "House BBQ Wings",
	14: "Mango-Habanero Wings",
	15: "Original Chicken Crispers®",
	16: "Southwestern Eggrolls",
}

// extras is a hash table that maps an Extra to its display name
var extras = map[Extra]string{
	0: "Ancho-Chile Ranch Dressing",
	1: "Avocado-Ranch Dressing",
	2: "Bleu Cheese Dressing",
	3: "Honey-Mustard Dressing",
	4: "Original BBQ Sauce",
	5: "Ranch Dressing",
}

// permitted is a hash table that maps an Item to its set of permitted Extras.
var permitted = map[Item][]Extra{
	0:  []Extra{1, 5},
	1:  []Extra{5},
	2:  []Extra{2, 5},
	3:  []Extra{2, 5},
	4:  []Extra{2, 5},
	5:  []Extra{2, 5},
	6:  []Extra{2, 5},
	7:  []Extra{0},
	8:  []Extra{3, 4, 5},
	9:  []Extra{3, 4, 5},
	10: []Extra{3, 4, 5},
	11: []Extra{1, 5},
	12: []Extra{2, 5},
	13: []Extra{2, 5},
	14: []Extra{2, 5},
	15: []Extra{3, 4, 5},
	16: []Extra{1, 5},
}

// An Item is the appetizer component of a Dipper.
type Item byte

// An Extra is an optional component of a Dipper, typically used for dipping
// sauces.
type Extra byte

// A Dipper is a component of a TripleDipper. It is composed of an Item and
// its Extras.
type Dipper struct {
	Item   Item
	Extras []Extra
}

type TripleDipper struct {
	Dippers [3]Dipper
}

// Name returns the Item's display name.
func (item Item) Name() string {
	return items[item]
}

// Permitted returns true if the given Extra is permitted for the Item.
func (item Item) Permitted(extra Extra) bool {
	for _, permitted := range permitted[item] {
		if extra == permitted {
			return true
		}
	}
	return false
}

// ParseID parses and returns an Item's Chili's ID given its selection index.
func (item Item) ParseID(node *html.Node, index int) (string, error) {
	text := fmt.Sprintf("Selection %d", index+1)
	label, err := findOne(node, textQuery("label", text))
	if err != nil {
		return "", fmt.Errorf("parsing Item's Chili's ID: %v", err)
	}
	selection := label.Parent
	option, err := findOne(selection, textQuery("option", item.Name()))
	if err != nil {
		return "", fmt.Errorf("parsing Item's Chili's ID: %v", err)
	}
	return htmlquery.SelectAttr(option, "value"), nil
}

// Name returns the Extra's display name.
func (extra Extra) Name() string {
	return extras[extra]
}

// ParseID parses and returns an Extra's Chili's ID given its Item's Chili's ID
func (extra Extra) ParseID(node *html.Node, itemID string) (string, error) {
	widgets, err := find(node, attrQuery("div", "data-related", itemID))
	if err != nil {
		return "", fmt.Errorf("parsing Extra's Chili's ID: %v", err)
	}
	for _, widget := range widgets {
		option, err := findOne(widget, textQuery("option", extra.Name()))
		if err == nil {
			return htmlquery.SelectAttr(option, "value"), nil
		}
	}
	return "", errors.New("parsing Extra's Chili's ID")
}

// Permitted returns true if all of the Dipper's Extras are permitted for the
// Dipper's item.
func (dipper Dipper) Permitted() bool {
	for _, extra := range dipper.Extras {
		if !dipper.Item.Permitted(extra) {
			return false
		}
	}
	return true
}
