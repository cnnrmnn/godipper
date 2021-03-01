package chilis

import (
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

// Name returns the Item's display name.
func (it Item) Name() string {
	return items[it]
}

// Permitted returns true if the given Extra is permitted for the Item.
func (it Item) Permitted(e Extra) bool {
	for _, p := range permitted[it] {
		if e == p {
			return true
		}
	}
	return false
}

// ParseID parses and returns an Item's Chili's ID given its selection index.
func (it Item) ParseID(node *html.Node, i int) (string, error) {
	text := fmt.Sprintf("Selection %d", i+1)
	label, err := findOne(node, textQuery("label", text))
	if err != nil {
		return "", fmt.Errorf("parsing Item's Chili's ID: %v", err)
	}
	opt, err := findOne(label.Parent, textQuery("option", it.Name()))
	if err != nil {
		return "", fmt.Errorf("parsing Item's Chili's ID: %v", err)
	}
	return htmlquery.SelectAttr(opt, "value"), nil
}
