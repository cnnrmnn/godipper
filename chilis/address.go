package chilis

import "fmt"

// An Address is a United States address.
type Address struct {
	Street string
	Unit   string
	City   string
	State  string
	Zip    string
}

// String returns a string representation of an Address (doesn't include Unit).
func (a Address) String() {
	return fmt.Sprintf("%s, %s, %s %s", a.Street, a.City, a.State, a.Zip)
}
