package chilis

// A Dipper is a component of a TripleDipper. It is composed of an Item and
// its Extras.
type Dipper struct {
	Item   Item
	Extras []Extra
}

// permitted returns true if all of the Dipper's Extras are permitted for the
// Dipper's item.
func (d Dipper) permitted() bool {
	for _, e := range d.Extras {
		if !d.Item.permitted(e) {
			return false
		}
	}
	return true
}
