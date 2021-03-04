package chilis

import (
	"testing"
)

var tests = []struct {
	extra Extra
	iids  []string
	ids   []string
}{
	{
		0,
		[]string{"285722848", "999898768", "284423526"},
		[]string{"285726142", "999901302", "284426814"},
	},
	{
		1,
		[]string{"285722846", "2267517501", "284423524"},
		[]string{"285726135", "2267518005", "284426807"},
	},
	{
		2,
		[]string{"1994890269", "1998403657", "1992203060"},
		[]string{"3226981473", "3226986236", "3226815555"},
	},
	{
		3,
		[]string{"285722847", "999898767", "284423525"},
		[]string{"285726137", "999901298", "284426809"},
	},
	{
		4,
		[]string{"285722843", "999898764", "284423521"},
		[]string{"1070970656", "1071223175", "1314697285"},
	},
	{
		5,
		[]string{"285722842", "999898763", "284423520"},
		[]string{"285726122", "999901286", "284426794"},
	},
}

func TestParseID(t *testing.T) {
	for n, doc := range dipperDocs {
		for _, test := range tests {
			path := dipperPaths[n]
			extra := test.extra
			id, err := extra.parseID(doc, test.iids[n])
			if err != nil {
				t.Errorf("%s (%d): %v", path, extra, err)
			}
			if id != test.ids[n] {
				t.Errorf("%s (%d): id = %s, want %s", path, extra, id, test.ids[n])
			}
		}
	}
}
