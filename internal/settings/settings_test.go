package settings

import "testing"

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(dir, "/dl")
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Update(func(s *Settings) { s.DeviceName = "Box"; s.Favorites["ab"] = "Phone" }); err != nil {
		t.Fatal(err)
	}
	st2, err := Open(dir, "/other")
	if err != nil {
		t.Fatal(err)
	}
	got := st2.Get()
	if got.DeviceName != "Box" || got.Favorites["ab"] != "Phone" || got.DownloadDir != "/dl" || got.Port != 53317 {
		t.Fatalf("%+v", got)
	}
}
