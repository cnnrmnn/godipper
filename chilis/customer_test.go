package chilis

import (
	"fmt"
	"testing"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

var paths = []string{
	"testdata/checkout1.html",
	"testdata/checkout2.html",
	"testdata/checkout3.html",
}
var docs []*html.Node

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

func init() {
	// Only read/parse test HTML files once
	for _, path := range paths {
		doc, err := htmlquery.LoadDoc(path)
		if err != nil {
			fmt.Errorf("%s: %v", path, err)
		}
		docs = append(docs, doc)
	}
}

func TestParseTotal(t *testing.T) {
	for n, test := range totalTests {
		path := paths[n]
		subtotal, tax, err := parseTotal(docs[n])
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
		path := paths[n]
		date, time, err := parseASAP(docs[n])
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
		path := paths[n]
		tid, err := parseTransactionID(docs[n])
		if err != nil {
			t.Errorf("%s: %v", path, err)
		}
		if test != tid {
			t.Errorf("%s: tid = %s, want %s", path, tid, test)
		}
	}
}
