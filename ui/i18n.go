package ui

import (
	"strconv"
	"time"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/i18n"

	"easysend/internal/transfer"
)

// tr translates a message. The English text is the key, so a string the
// locale files lack shows in English; {name} placeholders are filled from
// the name, value pairs that follow.
func tr(key string, kv ...any) string { return giode.Intl(key, kv...) }

// count translates a message with plural forms, chosen by n.
func count(key string, n int) string { return giode.Intl(key, "count", n) }

// clock formats a time of day.
func clock(t time.Time) string { return t.Format("15:04") }

// monthDay is a date without a year ("Jan 2"), with the month named in
// the current language.
func monthDay(t time.Time) string {
	return tr("{month} {day}", "month", tr("month."+strconv.Itoa(int(t.Month()))), "day", t.Day())
}

// weekdayTime is a day of this week and a time ("Mon 15:04").
func weekdayTime(t time.Time) string {
	return tr("{weekday} {time}", "weekday", tr("weekday."+strconv.Itoa(int(t.Weekday()))), "time", clock(t))
}

// dateTime is a date and a time ("Jan 2 15:04").
func dateTime(t time.Time) string {
	return tr("{date} {time}", "date", monthDay(t), "time", clock(t))
}

// scope names a cache of Keyed components for the current language, so
// their labels are rebuilt when the language changes.
func scope(name string) string { return name + "/" + i18n.Locale() }

// titleOf describes a transfer in a few words, in the current language.
func titleOf(t transfer.Transfer) string { return t.Title() }
