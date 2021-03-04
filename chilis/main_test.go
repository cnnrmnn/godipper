package chilis

import (
	"fmt"
	"os"
	"testing"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

var dipperPaths = []string{
	"testdata/dipper1.html",
	"testdata/dipper2.html",
	"testdata/dipper3.html",
}

var dipperDocs []*html.Node

func TestMain(m *testing.M) {
	// Only read/parse test HTML file once
	// Do here since shared by multiple test files
	for _, path := range dipperPaths {
		doc, err := htmlquery.LoadDoc(path)
		if err != nil {
			fmt.Errorf("%s: %v", path, err)
		}
		dipperDocs = append(dipperDocs, doc)
	}
	os.Exit(m.Run())
}
