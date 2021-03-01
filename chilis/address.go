package chilis

import "fmt"

// An Address is a United States address.
type Address struct {
	Street string `json:"street"`
	Unit   string `json:"unit"`
	City   string `json:"city"`
	State  string `json:"state"`
	Zip    string `json:"zip"`
}

// String returns a string representation of an Address (doesn't include Unit).
func (a Address) String() string {
	return fmt.Sprintf("%s, %s, %s %s", a.Street, a.City, a.State, a.Zip)
}

// Chilis retrns a string representation of an Address that is formatted for
// Chili's delivery estimate endpoint.
func (a Address) Chilis() string {
	return fmt.Sprintf("%s,%s,%s,%s", a.Street, a.City, a.State, a.Zip)
}
