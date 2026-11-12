package locales

import (
	"encoding/json"
	"regexp"
	"sort"
	"testing"
)

var placeholder = regexp.MustCompile(`\{\w+\}`)

func placeholders(raw json.RawMessage) []string {
	var forms map[string]string
	var text string
	var out []string
	if json.Unmarshal(raw, &text) == nil {
		out = placeholder.FindAllString(text, -1)
	} else if json.Unmarshal(raw, &forms) == nil {
		out = placeholder.FindAllString(forms["other"], -1)
	}
	sort.Strings(out)
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestTranslationsMatchEnglish checks that every language has every key,
// with the same placeholders, and no extra keys.
func TestTranslationsMatchEnglish(t *testing.T) {
	load := func(name string) map[string]json.RawMessage {
		b, err := FS.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		return m
	}
	en := load("en.json")
	files, _ := FS.ReadDir(".")
	for _, f := range files {
		if f.Name() == "en.json" {
			continue
		}
		m := load(f.Name())
		for k, v := range en {
			tv, ok := m[k]
			if !ok {
				t.Errorf("%s: missing %q", f.Name(), k)
			} else if !equal(placeholders(v), placeholders(tv)) {
				t.Errorf("%s: placeholders of %q differ", f.Name(), k)
			}
		}
		for k := range m {
			if _, ok := en[k]; !ok {
				t.Errorf("%s: extra key %q", f.Name(), k)
			}
		}
	}
}
