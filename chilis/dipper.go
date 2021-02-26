package chilis

// items is a hash table that maps an item to its display name.
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

// extras is a hash table that maps an extra to its display name
var extras = map[Extra]string{
	0: "Ancho-Chile Ranch Dressing",
	1: "Avocado-Ranch Dressing",
	2: "Bleu Cheese Dressing",
	3: "Honey-Mustard Dressing",
	4: "Original BBQ Sauce",
	5: "Ranch Dressing",
}

// permitted is a hash table that maps an item to its set of permitted extras.
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

// An Item is the appetizer component of a dipper.
type Item byte

// An Extra is an optional component of a dipper, typically used for dipping
// sauces.
type Extra byte

// A Dipper is a component of a triple dipper. It is composed of an item and
// its extras.
type Dipper struct {
	Item   Item
	Extras []Extra
}

type TripleDipper struct {
	Dippers [3]Dipper
}

// Name returns the item's display name.
func (item Item) Name() string {
	return items[item]
}

// Permitted returns true if the given extra is permitted for the item.
func (item Item) Permitted(extra Extra) bool {
	for _, permitted := range permitted[item] {
		if extra == permitted {
			return true
		}
	}
	return false
}

// Name returns the extra's display name.
func (extra Extra) Name() string {
	return extras[extra]
}

// Permitted returns true if all of the dipper's extras are permitted for the
// dipper's item.
func (dipper Dipper) Permitted() bool {
	for _, extra := range dipper.Extras {
		if !dipper.Item.Permitted(extra) {
			return false
		}
	}
	return true
}
