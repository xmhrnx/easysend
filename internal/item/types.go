// Package item describes the things a user sends: files (with an optional
// path inside a sent folder) and text snippets.
package item

// Kind tells a file from a text snippet.
type Kind int

const (
	// File is a file on disk.
	File Kind = iota
	// Text is a text snippet held in memory.
	Text
)

// Item is one thing to send.
type Item struct {
	ID string
	// Kind is File or Text.
	Kind Kind
	// Name is the name the receiver sees. For files inside a folder it is
	// the slash-separated path relative to the folder's parent
	// ("Photos/2024/a.jpg").
	Name string
	// Path is the file on disk (File only).
	Path string
	// Size in bytes.
	Size int64
	// MIME type.
	MIME string
	// Text is the snippet (Text only).
	Text string
	// Group names the folder an item was added with, for display.
	Group string
}
