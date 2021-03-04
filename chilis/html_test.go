package chilis

import "testing"

var tokenTests = []string{
	"9446659f-0635-4800-b48f-ecdbfe917fce",
	"5c557574-064e-4e78-8f36-4db08d07f260",
	"622d0a6c-aa6e-48c3-8498-babcde17711d",
}

func TestParseCSRFToken(t *testing.T) {
	for n, test := range tokenTests {
		path := dipperPaths[n]
		token, err := parseCSRFToken(dipperDocs[n])
		if err != nil {
			t.Errorf("%s: %v", path, err)
		}
		if token != test {
			t.Errorf("%s: token = %s, want %s", path, token, test)
		}
	}
}
