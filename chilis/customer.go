package chilis

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// A Customer provides access to all of the customer-related information that
// is needed to place a Chili's order. It isn't implemented as a concrete type
// (like all of the other types in the package) so that developers are free to
// define their own concrete types for users/customers that may contain more
// information than this package needs.
type Customer struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
}

type OrderInfo struct {
	Subtotal      string `json:"subtotal"`
	Tax           string `json:"tax"`
	DeliveryFee   string `json:"deliveryFee"`
	ServiceCharge string `json:"serviceCharge"`
	DeliveryTime  string `json:"deliveryTime"`
}

// checkoutForm adds all of the customer's information to a form map with the default
// values for every checkout request.
func checkoutForm(doc *html.Node, c Customer, addr Address) (url.Values, error) {
	form := url.Values{}
	form.Add("inAuthData.siteKey", "48693e4afc6b92d9")
	form.Add("inAuthData.collectorURL", "www.cdn-net.com")
	form.Add("inAuthData.collectorFlags", "34549755")
	form.Add("inAuthData.enabled", "true")
	form.Add("deliveryToggle", "on")
	form.Add("orderMode", "delivery")
	form.Add("deviceType", "web")
	form.Add("payment", "online")
	form.Add("silverwareOptIn", "true")
	form.Add("smsOptIn", "true")
	form.Add("deliveryAddress", addr.String())
	form.Add("deliveryAddress2", addr.Unit)
	form.Add("deliveryAddlNotes", addr.Notes)
	form.Add("firstName", c.FirstName)
	form.Add("lastName", c.LastName)
	form.Add("contactPhone", c.Phone)
	form.Add("email", c.Email)
	date, time, err := parseASAP(doc)
	if err != nil {
		return nil, fmt.Errorf("creating checkout form: %w", err)
	}
	// Chili's inexplicably requires all of these fields.
	form.Add("deliveryDate", date)
	form.Add("pickupDate", date)
	form.Add("deliveryTime", time)
	form.Add("pickupTime", time)
	tid, err := parseTransactionID(doc)
	if err != nil {
		return nil, fmt.Errorf("creating checkout form: %v", err)
	}
	form.Add("inAuthData.transactionId", tid)
	csrf, err := parseCSRFToken(doc)
	if err != nil {
		return nil, fmt.Errorf("creating checkout form: %v", err)
	}
	form.Add("_csrf", csrf)
	return form, nil
}

// validCustomer returns an error if any of the customer's methods return
// invalid values.
func validCustomer(c Customer) error {
	if err := validPhone(c.Phone); err != nil {
		return err
	}
	if err := validEmail(c.Email); err != nil {
		return err
	}
	return nil
}

// validPhone returns an error if the customer's phone isn't a string of ten
// digit runes.
func validPhone(phone string) error {
	n := 0
	for _, digit := range phone {
		if digit < '0' || digit > '9' {
			return BadRequestError{"phone"}
		}
		n++
	}
	if n != 10 {
		return BadRequestError{"phone"}
	}
	return nil
}

// validEmail returns an error if the customer's email doesn't have an @ rune.
func validEmail(email string) error {
	for _, r := range email {
		if r == '@' {
			return nil
		}
	}
	return BadRequestError{"email"}
}

func parseInfo(doc *html.Node) (info OrderInfo, err error) {
	info.Subtotal, err = innerText(doc, classQuery("div", "cost js-subtotal"))
	if err != nil {
		return info, fmt.Errorf("parsing subtotal: %v", err)
	}
	// XPath query
	q := "//tr[@id='pickup-tax-payment']/td[2]/div[@class='cost']"
	info.Tax, err = innerText(doc, q)
	if err != nil {
		return info, fmt.Errorf("parsing tax: %v", err)
	}
	// XPath query
	q = "//tr[@id='delivery-fee']/td[2]/div[@class='cost']"
	info.DeliveryFee, err = innerText(doc, q)
	if err != nil {
		return info, fmt.Errorf("parsing delivery fee: %v", err)
	}
	// XPath query
	q = "//tr[@id='service-charge']/td[2]/div[@class='cost']"
	info.ServiceCharge, err = innerText(doc, q)
	if err != nil {
		return info, fmt.Errorf("parsing service charge: %v", err)
	}
	return info, nil
}

// parseASAP parses and returns the ASAP values for the date and time fields
// in the checkout form.
func parseASAP(doc *html.Node) (string, string, error) {
	var date, time string
	q := attrQuery("div", "id", "delivery-time-group")
	con, err := findOne(doc, q)
	if err != nil {
		return date, time, fmt.Errorf("parsing ASAP delivery: %v", err)
	}
	// XPath query
	q = "/div/select[@id='delivery-date']/option"
	date, err = selectAttr(con, q, "value")
	if err != nil {
		return date, time, ForbiddenError{"location is not currently delivering"}
	}
	// XPath query
	q = "/div/select[@id='delivery-time']/option"
	time, err = selectAttr(con, q, "value")
	if err != nil {
		return date, time, fmt.Errorf("parsing ASAP delivery time: %v", err)
	}
	return date, time, nil
}

// parseTransactionID returns the transaction ID associated with the checkout
// form.
func parseTransactionID(doc *html.Node) (string, error) {
	var tid string
	input, err := findOne(doc, attrQuery("input", "id", "transactionId"))
	if err != nil {
		return tid, fmt.Errorf("parsing transaction ID: %v", err)
	}
	tid = htmlquery.SelectAttr(input, "value")
	return tid, nil
}

func parseEstimate(body []byte) (string, error) {
	var time string
	var decoded interface{}
	err := json.Unmarshal(body, &decoded)
	if err != nil {
		return time, fmt.Errorf("parsing delivery estimate body: %v", err)
	}
	time, ok := decoded.(map[string]interface{})["delivery_time"].(string)
	if !ok {
		return time, ForbiddenError{"address is out of range"}
	}
	return time, nil
}
