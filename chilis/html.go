package chilis

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// classQuery returns an XPath query for an HTML element with the given name and
// class attribute.
func classQuery(name string, class string) string {
	return attrQuery(name, "class", class)
}

// attrQuery returns an XPath query for an HTML element with the given name,
// attribute, and value for that attribute.
func attrQuery(name, attr, value string) string {
	return fmt.Sprintf("//%s[@%s='%s']", name, attr, value)
}

// textQuery returns an XPath query for an HTML element with the given name and
// inner text.
func textQuery(name, text string) string {
	return fmt.Sprintf("//%s[text()='%s']", name, text)
}

// find finds and returns a slice of HTML elements with the given name and
// class attribute.
func find(node *html.Node, query string) ([]*html.Node, error) {
	elements := htmlquery.Find(node, query)
	if elements == nil {
		return nil, errors.New("no matching elements found")
	}
	return elements, nil
}

// findOne is the same as find, but it only finds the first HTML element.
func findOne(node *html.Node, query string) (*html.Node, error) {
	element := htmlquery.FindOne(node, query)
	if element == nil {
		return nil, errors.New("no matching elements found")
	}
	return element, nil
}

// innerText finds the first HTML element with the given name and class
// attribute and returns its inner text.
func innerText(node *html.Node, query string) (string, error) {
	element, err := findOne(node, query)
	if err != nil {
		return "", fmt.Errorf("parsing inner text: %v", err)
	}
	innerText := htmlquery.InnerText(element)
	return innerText, nil
}

// parsePage parses and returns the root node of an HTML document at the given
// url
func parsePage(clt *http.Client, u string) (*html.Node, error) {
	resp, err := clt.Get(u)
	if err != nil {
		return nil, fmt.Errorf("fetching HTML at %s: %v", err)
	}
	defer resp.Body.Close()
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing HTML at %s: %v", err)
	}
	return doc, nil
}

// parseCSRFToken parses and returns the CSRF token given any Chili's form
// page.
func parseCSRFToken(node *html.Node) (string, error) {
	input, err := findOne(node, attrQuery("input", "name", "_csrf"))
	if err != nil {
		return "", fmt.Errorf("parsing CSRF token: %v", err)
	}
	return htmlquery.SelectAttr(input, "value"), nil
}
