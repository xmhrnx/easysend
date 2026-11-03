package transfer

import (
	"testing"
)

func TestRegistry(t *testing.T) {
	dir := t.TempDir()
	r := NewRegistry(dir, nil)
	id := r.Start(Transfer{Method: Nearby, Files: []File{{Name: "a", Size: 10}, {Name: "b", Size: 30}}}, nil)
	r.AddBytes(id, 0, 10)
	got, _ := r.Get(id)
	if got.Total != 40 || got.Done != 10 || got.Status != Active || got.Files[0].Done != 10 {
		t.Fatalf("%+v", got)
	}
	r.Finish(id, Done, nil)
	r2 := NewRegistry(dir, nil)
	list := r2.List()
	if len(list) != 1 || list[0].Status != Done || list[0].Progress() != 1 {
		t.Fatalf("%+v", list)
	}
	if list[0].Title() != "2 files" {
		t.Fatal(list[0].Title())
	}
}
