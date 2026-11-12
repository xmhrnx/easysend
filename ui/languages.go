package ui

import (
	"slices"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/i18n"
)

// language is a translation EasySend ships, named in its own script.
type language struct {
	Code string
	Name string
}

// languages are the interface languages, in the order the picker lists them.
var languages = []language{
	{"en", "English"},
	{"ka", "ქართული"},
	{"hy", "Հայերեն"},
	{"he", "עברית"},
	{"ar", "العربية"},
	{"fa", "فارسی"},
	{"ru", "Русский"},
	{"uk", "Українська"},
	{"bg", "Български"},
	{"el", "Ελληνικά"},
	{"de", "Deutsch"},
	{"fr", "Français"},
	{"es", "Español"},
	{"pt", "Português"},
	{"it", "Italiano"},
	{"nl", "Nederlands"},
	{"sv", "Svenska"},
	{"da", "Dansk"},
	{"nb", "Norsk"},
	{"fi", "Suomi"},
	{"pl", "Polski"},
	{"cs", "Čeština"},
	{"hu", "Magyar"},
	{"ro", "Română"},
	{"tr", "Türkçe"},
	{"hi", "हिन्दी"},
	{"id", "Bahasa Indonesia"},
	{"vi", "Tiếng Việt"},
	{"th", "ไทย"},
	{"zh", "中文"},
	{"ja", "日本語"},
	{"ko", "한국어"},
}

// languageOptions lists the picker's entries: the system language first,
// then every language by its own name.
func languageOptions() []string {
	opts := []string{tr("System language")}
	for _, l := range languages {
		opts = append(opts, l.Name)
	}
	return opts
}

// languageIndex is the picker position of a saved language code; 0 (the
// system language) when the code is empty or unknown.
func languageIndex(code string) int {
	i := slices.IndexFunc(languages, func(l language) bool { return l.Code == code })
	return i + 1
}

// languageCode is the code for a picker position; "" for the system language.
func languageCode(index int) string {
	if index < 1 || index > len(languages) {
		return ""
	}
	return languages[index-1].Code
}

// ApplyLanguage switches the interface to a saved language code, or back
// to the system's language when the code is empty.
func ApplyLanguage(code string) {
	if code == "" {
		code = i18n.SystemLocale()
	}
	giode.SetLocale(code)
}
