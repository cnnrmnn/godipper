package chilis

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type Customer struct {
	Address   Address `json:"address"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Phone     string  `json:"phone"`
	Email     string  `json:"email"`
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
	date, time, err := parseASAP(doc)
	if err != nil {
		return nil, fmt.Errorf("creating checkout form: %v", err)
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
	return form, nil
}

// deliveryTime returns an estimated delivery time or an error if the customer's
// address is out of range.
func (c Customer) deliveryTime(clt *http.Client, csrf string) (t time.Time, err error) {
	u := "https://www.chilis.com/order/delivery/estimate"
	form := url.Values{}
	form.Add("_csrf", csrf)
	// Delivery estimate form requires this strange address format.
	form.Add("deliveryAddress", c.Address.Chilis())
	resp, err := clt.PostForm(u, form)
	if err != nil {
		return t, fmt.Errorf("fetching delivery estimate: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return t, fmt.Errorf("reading delivery estimate response: %v", err)
	}
	defer resp.Body.Close()

	var estimate *map[string]string
	err = json.Unmarshal(body, estimate)
	if err != nil {
		return t, fmt.Errorf("parsing delivery estimate response: %v", err)
	}
	tstr, ok := (*estimate)["delivery_time"]
	if !ok {
		return t, errors.New("address is out of range")
	}
	t, err = time.Parse(time.RFC3339, tstr)
	if err != nil {
		return t, fmt.Errorf("parsing delivery time estimate: %v", err)
	}
	return t, nil
}

// valid returns true if the customer's phone and email are valid.
func (c Customer) valid() bool {
	return validPhone(c.Phone) && validEmail(c.Email)
}

// parseTotal returns a map with the order's subtotal and estimated tax
func parseTotal(doc *html.Node) (string, string, error) {
	subtotal, err := innerText(doc, classQuery("div", "cost js-subtotal"))
	if err != nil {
		return "", "", fmt.Errorf("parsing subtotal: %v", err)
	}
	// Slightly complex query in raw XPath
	q := "//tr[@id='pickup-tax-payment]/td[2]/div"
	tax, err := innerText(doc, q)
	if err != nil {
		return "", "", fmt.Errorf("parsing tax: %v", err)
	}
	return subtotal, tax, nil
}

// parseASAP parses and returns the ASAP values for the date and time fields
// in the checkout form.
func parseASAP(doc *html.Node) (date, time string, err error) {
	con, err := findOne(doc, classQuery("div", "delivery-time-group"))
	if err != nil {
		return date, time, fmt.Errorf("parsing ASAP delivery: %v", err)
	}
	// Slightly complicated XPath query
	dq := "/div/select[@id='delivery-date']/option"
	dopt, err := findOne(con, dq)
	if err != nil {
		return date, time, fmt.Errorf("parsing ASAP delivery date: %v", err)
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
	input, err := findOne(doc, attrQuery("input", "id", "transactionId"))
	if err != nil {
		return "", fmt.Errorf("parsing transaction ID: %v", err)
	}
	return htmlquery.SelectAttr(input, "value"), nil
}

// validPhone returns true if the given string has 10 digit runes.
func validPhone(phone string) bool {
	n := 0
	for _, digit := range phone {
		if digit < '0' || digit > '9' {
			return false
		}
		n++
	}
	if n != 10 {
		return false
	}
	return true
}

// validEmail returns true if the given string has at least one @ rune.
func validEmail(email string) bool {
	for _, c := range email {
		if c == '@' {
			return true
		}
	}
	return false
}
