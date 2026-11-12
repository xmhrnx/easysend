package ui

import (
	"path"
	"strings"
	"time"

	"easysend/internal/item"
	"easysend/internal/localsend"
	"easysend/internal/transfer"
)

// deviceIcon names the icon for a LocalSend device type.
func deviceIcon(deviceType string) string {
	switch deviceType {
	case localsend.TypeMobile:
		return "phone"
	case localsend.TypeWeb:
		return "globe"
	case localsend.TypeServer, localsend.TypeHeadless:
		return "server"
	}
	return "laptop"
}

// fileIcon names the icon for a file name.
func fileIcon(name string) string {
	mime := item.MIMEOf(name)
	switch {
	case strings.HasPrefix(mime, "image/"):
		return "image"
	case strings.HasPrefix(mime, "video/"):
		return "video"
	case strings.HasPrefix(mime, "audio/"):
		return "music"
	case strings.HasPrefix(mime, "text/"):
		return "text"
	}
	switch strings.ToLower(path.Ext(name)) {
	case ".zip", ".rar", ".7z", ".tar", ".gz", ".xz":
		return "archive"
	}
	return "file"
}

func methodLabel(m transfer.Method) string {
	switch m {
	case transfer.Internet:
		return tr("Internet")
	case transfer.Browser:
		return tr("Browser")
	}
	return tr("Nearby")
}

func methodIcon(m transfer.Method) string {
	switch m {
	case transfer.Internet:
		return "globe"
	case transfer.Browser:
		return "link"
	}
	return "wifi"
}

// timeAgo formats a past time ("just now", "5 min ago", "Mar 3").
func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return tr("just now")
	case d < time.Hour:
		return count("minutes_ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return count("hours_ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return weekdayTime(t)
	}
	return monthDay(t)
}
