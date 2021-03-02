package chilis

import (
	"errors"
	"fmt"

	creditcard "github.com/durango/go-credit-card"
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

// Validate verifies that the payment method has a valid number and adds the
// company to the payment method.
func (pm *PaymentMethod) Validate() error {
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
func (pm *PaymentMethod) Format() (f string, err error) {
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
