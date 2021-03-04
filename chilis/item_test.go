package chilis

import "testing"

var itemTests = []struct {
	item Item
	ids  []string
}{
	{0, []string{"1569900708", "2067630827", "1568797388"}},
	{1, []string{"285722842", "999898763", "284423520"}},
	{2, []string{"285722851", "999898771", "284423529"}},
	{3, []string{"285722849", "999898769", "284423527"}},
	{4, []string{"1994890268", "1998403656", "1992203059"}},
	{5, []string{"3418702265", "3424231160", "3417912673"}},
	{6, []string{"285722838", "999898761", "284423516"}},
	{7, []string{"285722848", "999898768", "284423526"}},
	{8, []string{"285722847", "999898767", "284423525"}},
	{9, []string{"285722850", "999898770", "284423528"}},
	{10, []string{"3418702264", "3424231159", "3417912672"}},
	{11, []string{"285722846", "2267517501", "284423524"}},
	{12, []string{"285722844", "999898765", "284423522"}},
	{13, []string{"1994890269", "1998403657", "1992203060"}},
	{14, []string{"3418702266", "3424231161", "3417912674"}},
	{15, []string{"285722843", "999898764", "284423521"}},
	{16, []string{"285722852", "999898772", "284423530"}},
}

// Test each item as the nth selection in dipper<n>.html
func TestParseIDItem(t *testing.T) {
	for n, doc := range dipperDocs {
		for _, test := range itemTests {
			path := dipperPaths[n]
			item := test.item
			id, err := item.parseID(doc, n)
			if err != nil {
				t.Errorf("%s (%d): %v", path, item, err)
			}
			if id != test.ids[n] {
				t.Errorf("%s (%d): id = %s, want %s", path, item, id, test.ids[n])
			}
		}
	}
}
