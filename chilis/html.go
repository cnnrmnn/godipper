package chilis

import (
	"errors"
	"fmt"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// classQuery returns an XPath query for an HTML element with the given name and
// class attribute.
func classQuery(name string, class string) string {
	return fmt.Sprintf("//%s[@class='%s']", name, class)
}

// innerText finds the first HTML element with the given name and class
// attribute and returns its inner text.
func innerText(node *html.Node, name string, class string) (string, error) {
	query := classQuery(name, class)
	tag := htmlquery.FindOne(node, query)
	if tag == nil {
		return "", errors.New("no matching elements found")
	}
	innerText := htmlquery.InnerText(tag)
	return innerText, nil
}
