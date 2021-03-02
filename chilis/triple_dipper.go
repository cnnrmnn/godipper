package chilis

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

type TripleDipper struct {
	Dippers [3]Dipper
}

// Cart adds the TripleDipper to the cart given a session cookie.
func (td TripleDipper) Cart(clt *http.Client) error {
	u := "https://www.chilis.com/menu/appetizers/triple-dipper"
	doc, err := parsePage(clt, u)
	if err != nil {
		return fmt.Errorf("adding TripleDipper to cart: %v", err)
	}

	form, err := td.form(doc)
	if err != nil {
		return fmt.Errorf("adding TripleDipper to cart: %w", err)
	}

	resp, err := clt.PostForm(u, form)
	if err != nil {
		return fmt.Errorf("posting cart request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading cart response body: %v", err)
	}
	// Request was successful if response body is a valid JSON encoding
	if !json.Valid(body) {
		return errors.New("unable to add TripleDipper to cart")
	}
	return nil
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

		iid, err := d.Item.ParseID(doc, i)
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
