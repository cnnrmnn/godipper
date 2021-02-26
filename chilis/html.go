package chilis

import (
	"errors"
	"fmt"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// findOne finds and returns the first HTML element with the given name and
// class attribute
func findOne(node *html.Node, name, class string) (node *html.Node, error) {
	query := fmt.Sprintf("//%s[@class='%s']", name, class)
	element := htmlquery.FindOne(node, query)
	if (element == nil) {
		return nil, errors.New("no matching elements found")
	}
	return element, nil
}

// innerText finds the first HTML element with the given name and class
// attribute and returns its inner text.
func innerText(node *html.Node, name, class string) (string, error) {
	element, err := findOne(node, name, class)
	if (err != nil) {
		return "", fmt.Errorf("failed to get inner text: %v", err)
	}
	innerText := htmlquery.InnerText(element)
	return innerText, nil
}
