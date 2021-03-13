package chilis

import "fmt"

// An Address is a United States address.
type Address struct {
	Street string `json:"street"`
	Unit   string `json:"unit"`
	City   string `json:"city"`
	State  string `json:"state"`
	Zip    string `json:"zip"`
	Notes  string `json:"notes"`
}

// String retrns a string representation of an Address that is formatted for
// Chili's delivery endpoints.
func (a Address) String() string {
	return fmt.Sprintf("%s,%s,%s,USA", a.Street, a.City, a.State)
}
