package chilis

import (
	"errors"
	"fmt"
	"net/http"
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

func (pm *PaymentMethod) Order(sid string) (Location, error) {
	var loc Location
	err := pm.validate()
	if err != nil {
		return loc, fmt.Errorf("creating order: %v", err)
	}
	u := "https://www.chilis.com/order/payment"
	session := http.Cookie{Name: "SESSION", Value: sid}
	clt, err := createClient(&session)
	if err != nil {
		return loc, fmt.Errorf("creating order client: %v", err)
	}
	doc, err := parsePage(clt, u)
	if err != nil {
		return loc, fmt.Errorf("fetching payment information: %v", err)
	}
	form, err := pm.form(doc)
	if err != nil {
		return loc, fmt.Errorf("bulding order request: %v", err)
	}
	resp, err := clt.PostForm(u, form)
	if err != nil {
		return loc, fmt.Errorf("posting order request: %v", err)
	}
	defer resp.Body.Close()
	doc, err = html.Parse(resp.Body)
	if err != nil {
		return loc, fmt.Errorf("parsing order response: %v", err)
	}
	loc, err = parseLocation(doc)
	if err != nil {
		return loc, fmt.Errorf("parsing order response: %v", err)
	}
	return loc, err
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
	err := card.Validate(true)
	if err != nil {
		return err
	}
	err = card.Method()
	if err != nil {
		return err
	}
	s := card.Company.Short
	switch s {
	case "visa", "mastercard", "discover":
		pm.Company = s
	case "amex":
		pm.Company = "americanexpress"
	default:
		return errors.New("invalid card company")
	}
	return nil
}

// Format returns the card number formatted in Chili's style according to its
// company. Validate must be called prior to Format or it will fail.
func (pm *PaymentMethod) format() (f string, err error) {
	c := pm.Number
	switch pm.Company {
	case "visa", "mastercard", "discover":
		f = fmt.Sprintf("%s-%s-%s-%s", c[:4], c[4:8], c[8:12], c[12:])
	case "americanexpress":
		f = fmt.Sprintf("%s-%s-%s", c[:4], c[4:10], c[10:])
	default:
		return "", errors.New("invalid or unset card company")
	}
	return f, nil
}
