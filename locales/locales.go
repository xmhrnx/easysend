// Package locales holds EasySend's translations: one JSON file per
// language, mapping the English text (or a plural key) to its translation.
package locales

import "embed"

// FS holds the message files, one per language code.
//
//go:embed *.json
var FS embed.FS
