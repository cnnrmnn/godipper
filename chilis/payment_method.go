package chilis

import (
	"errors"
	"fmt"
	"net/url"

	creditcard "github.com/durango/go-credit-card"
	"golang.org/x/net/html"
)

// A PaymentMethod contains credit card data needed to submit an order.
type PaymentMethod struct {
	Number  string `json:"number"`
	CVV     string `json:"cvv"`
	Month   string `json:"month`
	Year    string `json:"year"`
	Name    string `json:"name"`
	Zip     string `json:"zip`
	Company string `json:"company"`
}

// form validates the payment method and adds all of its fields to a form map
// with default values set.
func (pm *PaymentMethod) form(doc *html.Node) (url.Values, error) {
	form := url.Values{}
	form.Add("paymentMethod", "creditcard")
	form.Add("orderMode", "delivery")
	form.Add("cardType", pm.Company)
	form.Add("cvv", pm.CVV)
	form.Add("expirationMonth", pm.Month)
	form.Add("expirationYear", pm.Year)
	form.Add("nameOnCard", pm.Name)
	form.Add("zipcode", pm.Zip)
	number, err := pm.format()
	if err != nil {
		return nil, fmt.Errorf("formatting card number: %v", err)
	}
	form.Add("cardNumber", number)
	csrf, err := parseCSRFToken(doc)
	if err != nil {
		return nil, fmt.Errorf("creating order form: %v", err)
	}
	form.Add("_csrf", csrf)
	return form, nil
}

// Validate verifies that the payment method has a valid number and adds the
// company to the payment method.
func (pm *PaymentMethod) validate() error {
	card := creditcard.Card{
		Number: pm.Number,
		Cvv:    pm.CVV,
		Month:  pm.Month,
		Year:   pm.Year,
	}

	if ok := card.ValidateNumber(); !ok {
		return BadRequestError{"credit card number"}
	}
	if err := card.ValidateExpiration(); err != nil {
		return BadRequestError{"expiration date"}
	}
	if err := card.ValidateCVV(); err != nil {
		return BadRequestError{"cvv"}
	}

	if err := card.Method(); err != nil {
		return err
	}
	s := card.Company.Short
	switch s {
	case "visa", "mastercard", "discover":
		pm.Company = s
	case "amex":
		pm.Company = "americanexpress"
	default:
		return BadRequestError{"credit card company"}
	}
	return nil
}

// Format returns the card number formatted in Chili's style according to its
// company. Validate must be called prior to Format or it will fail.
func (pm *PaymentMethod) format() (string, error) {
	var f string
	c := pm.Number
	switch pm.Company {
	case "visa", "mastercard", "discover":
		f = fmt.Sprintf("%s-%s-%s-%s", c[:4], c[4:8], c[8:12], c[12:])
	case "americanexpress":
		f = fmt.Sprintf("%s-%s-%s", c[:4], c[4:10], c[10:])
	default:
		return f, errors.New("invalid or unset card company")
	}
	return f, nil
}
