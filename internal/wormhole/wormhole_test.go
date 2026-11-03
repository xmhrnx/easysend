package wormhole

import (
	"testing"

	"easysend/internal/item"
)

func TestNormalizeCode(t *testing.T) {
	for in, want := range map[string]string{
		" 7 Guitarist Revenge ":          "7-guitarist-revenge",
		"wormhole-transfer:3-able-baker": "3-able-baker",
	} {
		if got := NormalizeCode(in); got != want {
			t.Errorf("%q: %q", in, got)
		}
	}
}

func TestEntries(t *testing.T) {
	items := []item.Item{{Name: "Pics/a.jpg", Group: "Pics"}, {Name: "Pics/s/b.jpg", Group: "Pics"}}
	if folderName(items) != "Pics" {
		t.Fatal(folderName(items))
	}
	e := entries(items)
	if e[0].Path != "a.jpg" || e[1].Path != "s/b.jpg" {
		t.Fatal(e)
	}
	mixed := []item.Item{{Name: "a.txt"}, {Name: "Text.txt"}}
	if folderName(mixed) != "EasySend (2 files)" || entries(mixed)[0].Path != "a.txt" {
		t.Fatal(folderName(mixed))
	}
}
