package ui

// navEntry describes one destination in the navigation.
type navEntry struct {
	screen screen
	label  string
	icon   string
}

var navEntries = []navEntry{
	{screenSend, "Send", "send"},
	{screenReceive, "Receive", "inbox"},
	{screenActivity, "Activity", "history"},
	{screenSettings, "Settings", "settings"},
}

// badge returns the count shown next to a destination.
func (u *UI) badge(s screen) int {
	switch s {
	case screenReceive:
		return len(u.eng.Requests())
	case screenActivity:
		n := 0
		for _, t := range u.eng.Transfers.List() {
			if !t.Status.Finished() {
				n++
			}
		}
		return n
	}
	return 0
}
