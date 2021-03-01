package chilis

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/html"
)

type Customer struct {
	Address   Address `json:"address"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Phone     string  `json:"phone"`
	Email     string  `json:"email"`
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
func parseTotal(doc *html.Node) (map[string]string, error) {
	subtotal, err := innerText(doc, classQuery("div", "cost js-subtotal"))
	if err != nil {
		return nil, fmt.Errorf("parsing subtotal: %v", err)
	}
	// Slightly complex query in raw XPath
	q := "//tr[@id='pickup-tax-payment]/td[2]/div"
	tax, err := innerText(doc, q)
	return map[string]string{
		"subtotal": subtotal,
		"tax":      tax,
	}, nil
}

// defaultForm returns a map with values used for every checkout request.
func defaultForm() (form url.Values) {
	// Writing out string slice literals is annoying.
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
	return form
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
