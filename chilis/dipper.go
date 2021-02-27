package chilis

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

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

// parsePage parses the Triple Dipper page and returns its root node given a
// session cookie.
func parsePage(client *http.Client, url string) (*html.Node, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching Triple Dipper page: %v", err)
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing Triple Dipper page: %v", err)
	}
	return doc, nil
}

// form checks if the TripleDipper is permitted and adds all of its component's
// Chili's IDs and a CSRF token to the given form Values map.
func (tripleDipper TripleDipper) form(doc *html.Node) (*url.Values, error) {
	form := &url.Values{}
	csrfToken, err := parseCSRFToken(doc)
	if err != nil {
		return nil, fmt.Errorf("adding CSRF token to cart: %v", err)
	}

	form.Add("_csrf", csrfToken)
	for i, dipper := range tripleDipper.Dippers {
		if !dipper.Permitted() {
			return nil, fmt.Errorf("dipper %d is not permitted", i)
		}

		itemID, err := dipper.Item.ParseID(doc, i)
		if err != nil {
			return nil, fmt.Errorf("adding Item to cart: %v", err)
		}
		form.Add("selectedIds", itemID)

		for _, extra := range dipper.Extras {
			extraID, err := extra.ParseID(doc, itemID)
			if err != nil {
				return nil, fmt.Errorf("adding Extra to cart: %v", err)
			}
			form.Add("selectedIds", extraID)
		}
	}
	return form, nil
}

// Cart adds the TripleDipper to the cart given a session cookie.
func (tripleDipper TripleDipper) Cart(session *http.Cookie) error {
	jar, err := createJar(session)
	if err != nil {
		fmt.Errorf("creating CookieJar for cart request: %v", err)
	}
	client := &http.Client{Jar: jar}

	u := "https://www.chilis.com/menu/appetizers/triple-dipper"
	doc, err := parsePage(client, u)
	if err != nil {
		return fmt.Errorf("adding TripleDipper to cart: %v", err)
	}

	form, err := tripleDipper.form(doc)
	if err != nil {
		return fmt.Errorf("adding TripleDipper to cart: %v", err)
	}

	resp, err := client.PostForm(u, *form)
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
