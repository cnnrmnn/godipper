package chilis

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"golang.org/x/net/html"
)

// TripleDipper is a Chili's triple dipper.
type TripleDipper interface {
	ItemValues() []Item
}

// form checks if the TripleDipper is permitted and adds all of its components'
// Chili's IDs and a CSRF token to the given form Values map.
func tripleDipperForm(doc *html.Node, td TripleDipper) (url.Values, error) {
	form := url.Values{}
	csrf, err := parseCSRFToken(doc)
	if err != nil {
		return nil, fmt.Errorf("adding CSRF token to form: %v", err)
	}
	form.Add("_csrf", csrf)

	for i, it := range td.ItemValues() {
		iid, err := parseItemID(doc, it.String(), i)
		if err != nil {
			return nil, fmt.Errorf("adding Item to form: %v", err)
		}
		form.Add("selectedIds", iid)

		for _, e := range it.ExtraValues() {
			eid, err := parseExtraID(doc, e, iid)
			if err != nil {
				return nil, fmt.Errorf("adding Extra to form: %v", err)
			}
			form.Add("selectedIds", eid)
		}
	}
	return form, nil
}

func parseCart(body []byte) error {
	var decoded interface{}
	err := json.Unmarshal(body, &decoded)
	if err != nil {
		return errors.New("parsing cart response body")
	}
	_, ok := decoded.(map[string]interface{})["error"]
	if ok {
		return errors.New("can't add invalid item to cart")
	}
	return nil

}
