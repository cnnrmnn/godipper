package chilis

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"golang.org/x/net/html"
)

type TripleDipper struct {
	Dippers [3]Dipper
}

// form checks if the TripleDipper is permitted and adds all of its components'
// Chili's IDs and a CSRF token to the given form Values map.
func (td TripleDipper) form(doc *html.Node) (url.Values, error) {
	form := url.Values{}
	csrf, err := parseCSRFToken(doc)
	if err != nil {
		return nil, fmt.Errorf("adding CSRF token to form: %v", err)
	}
	form.Add("_csrf", csrf)

	for i, d := range td.Dippers {
		if !d.permitted() {
			return nil, BadRequestError{fmt.Sprintf("dipper %d", i+1)}
		}

		iid, err := d.Item.parseID(doc, i)
		if err != nil {
			return nil, fmt.Errorf("adding Item to form: %v", err)
		}
		form.Add("selectedIds", iid)

		for _, e := range d.Extras {
			eid, err := e.parseID(doc, iid)
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
