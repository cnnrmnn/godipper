package chilis

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type Customer struct {
	Address   Address `json:"address"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Phone     string  `json:"phone"`
	Email     string  `json:"email"`
	Notes     string  `json:"notes"`
}

// Checkout submits the customer's information and returns a map with the order
// subtotal, tax, and estimated delivery time.
func (c Customer) Checkout(clt *http.Client) (map[string]string, error) {
	if err := c.valid(); err != nil {
		return nil, err
	}

	u := "https://www.chilis.com/order/pickup"
	doc, err := parsePage(clt, u)
	if err != nil {
		return nil, fmt.Errorf("fetching delivery information: %v", err)
	}

	subtotal, tax, err := parseTotal(doc)
	if err != nil {
		return nil, fmt.Errorf("parsing order total: %v", err)
	}

	form, err := c.form(doc)
	if err != nil {
		return nil, fmt.Errorf("building checkout request: %w", err)
	}

	time, err := c.deliveryEstimate(clt, form.Get("_csrf"))
	if err != nil {
		return nil, fmt.Errorf("getting delivery time: %w", err)
	}

	_, err = clt.PostForm(u, form)
	if err != nil {
		return nil, fmt.Errorf("posting checkout request: %v", err)
	}

	return map[string]string{
		"subtotal":     subtotal,
		"tax":          tax,
		"deliveryTime": time,
	}, nil
}

// form adds all of the customer's information to a form map with the default
// values for every checkout request
func (c Customer) form(doc *html.Node) (url.Values, error) {
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
	form.Add("deliveryAddress", c.Address.String())
	form.Add("deliveryAddress2", c.Address.Unit)
	form.Add("firstName", c.FirstName)
	form.Add("lastName", c.LastName)
	form.Add("contactPhone", c.Phone)
	form.Add("email", c.Email)
	form.Add("deliveryAddlNotes", c.Notes)
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

// deliveryEstimate returns an estimated delivery time or an error if the customer's
// address is out of range.
func (c Customer) deliveryEstimate(clt *http.Client, csrf string) (string, error) {
	var time string
	u := "https://www.chilis.com/order/delivery/estimate"
	form := url.Values{}
	form.Add("_csrf", csrf)
	// Delivery estimate form requires this strange address format.
	form.Add("deliveryAddress", c.Address.chilis())
	resp, err := clt.PostForm(u, form)
	if err != nil {
		return time, fmt.Errorf("fetching delivery estimate: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return time, fmt.Errorf("reading delivery estimate response: %v", err)
	}

	return parseEstimate(body)
}

// valid returns an error if any of the customer's fields are invalid.
func (c Customer) valid() error {
	if err := c.validPhone(); err != nil {
		return err
	}
	if err := c.validEmail(); err != nil {
		return err
	}
	return nil
}

// validPhone returns an error if the customer's phone isn't a string of ten
// digit runes.
func (c Customer) validPhone() error {
	n := 0
	for _, digit := range c.Phone {
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
func (c Customer) validEmail() error {
	for _, r := range c.Email {
		if r == '@' {
			return nil
		}
	}
	return BadRequestError{"email"}
}

// parseTotal returns a map with the order's subtotal and estimated tax
func parseTotal(doc *html.Node) (string, string, error) {
	var subtotal, tax string
	subtotal, err := innerText(doc, classQuery("div", "cost js-subtotal"))
	if err != nil {
		return subtotal, tax, fmt.Errorf("parsing subtotal: %v", err)
	}
	// Slightly complex query in raw XPath
	q := "//tr[@id='pickup-tax-payment']/td[2]/div"
	tax, err = innerText(doc, q)
	if err != nil {
		return subtotal, tax, fmt.Errorf("parsing tax: %v", err)
	}
	return subtotal, tax, nil
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
	// Slightly complicated XPath query
	dq := "/div/select[@id='delivery-date']/option"
	dopt, err := findOne(con, dq)
	if err != nil {
		return date, time, ForbiddenError{"location is not currently delivering"}
	}
	date = htmlquery.SelectAttr(dopt, "value")
	// Slightly complicated XPath query
	tq := "/div/select[@id='delivery-time']/option"
	topt, err := findOne(con, tq)
	if err != nil {
		return date, time, fmt.Errorf("parsing ASAP delivery time: %v", err)
	}
	time = htmlquery.SelectAttr(topt, "value")
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
