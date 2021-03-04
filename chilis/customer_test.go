package chilis

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

var checkoutPaths = []string{
	"testdata/checkout1.html",
	"testdata/checkout2.html",
	"testdata/checkout3.html",
}
var checkoutDocs []*html.Node

var totalTests = []struct {
	subtotal string
	tax      string
}{
	{"$13.19", "$0.93"},
	{"$40.47", "$2.43"},
	{"$83.64", "$6.90"},
}

var asapTests = []struct {
	date string
	time string
}{
	{"[ASAP] 20210303 14:07", "20210303 14:15"},
	{"[ASAP] 20210303 15:03", "20210303 15:15"},
	{"[ASAP] 20210303 12:02", "20210303 12:15"},
}

var tidTests = []string{
	"fe1c8b07-d8f5-4bb6-91c0-dfe497fa0ecf",
	"8d772a03-3015-477e-bdfc-c0f147ec393f",
	"61a205ae-98f5-48af-98ba-7471e74907be",
}

var estimateTests = []struct {
	path string
	time string
}{
	{"testdata/estimate1.json", "2021-03-03T21:37:43.258000Z"},
	{"testdata/estimate2.json", "2021-03-03T21:42:26.730000Z"},
	{"testdata/estimate3.json", "2021-03-03T21:58:00.234000Z"},
}

func init() {
	// Only read/parse test HTML files once
	for _, path := range checkoutPaths {
		doc, err := htmlquery.LoadDoc(path)
		if err != nil {
			fmt.Errorf("%s: %v", path, err)
		}
		checkoutDocs = append(checkoutDocs, doc)
	}
}

func TestParseTotal(t *testing.T) {
	for n, test := range totalTests {
		path := checkoutPaths[n]
		subtotal, tax, err := parseTotal(checkoutDocs[n])
		if err != nil {
			t.Errorf("%s: %v", path, err)
		}
		if subtotal != test.subtotal {
			t.Errorf("%s: subtotal = %s, want %s", path, subtotal, test.subtotal)
		}
		if tax != test.tax {
			t.Errorf("%s: tax = %s, want %s", path, tax, test.tax)
		}
	}
}

func TestParseASAP(t *testing.T) {
	for n, test := range asapTests {
		path := checkoutPaths[n]
		date, time, err := parseASAP(checkoutDocs[n])
		if err != nil {
			t.Errorf("%s: %v", path, err)
		}
		if date != test.date {
			t.Errorf("%s: date = %s, want %s", path, date, test.date)
		}
		if time != test.time {
			t.Errorf("%s: time = %s, want %s", path, time, test.time)
		}
	}
}

func TestParseTransactionID(t *testing.T) {
	for n, test := range tidTests {
		path := checkoutPaths[n]
		tid, err := parseTransactionID(checkoutDocs[n])
		if err != nil {
			t.Errorf("%s: %v", path, err)
		}
		if test != tid {
			t.Errorf("%s: tid = %s, want %s", path, tid, test)
		}
	}
}

func TestParseEstimate(t *testing.T) {
	for _, test := range estimateTests {
		path := test.path
		body, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("%s: %v", path, err)
		}
		time, err := parseEstimate(body)
		if err != nil {
			t.Errorf("%s: %v", path, err)
		}
		if time != test.time {
			t.Errorf("%s: time = %s, want %s", path, time, test.time)
		}
	}
}

func TestParseEstimateRange(t *testing.T) {
	// No need to do more than one
	path := "testdata/estimater1.json"
	reason := "address is out of range"
	body, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("%s: %v", path, err)
	}
	_, err = parseEstimate(body)
	if err == nil {
		t.Errorf("%s: err = nil, want %s", path, reason)
	}
	var e ForbiddenError
	if !errors.As(err, &e) {
		t.Errorf("%s: got err = %v, want (ForbiddenError) %s", path, err, reason)
	}
	if reason != e.Reason {
		t.Errorf("%s: err = %v, want %s", path, err, reason)
	}
}
