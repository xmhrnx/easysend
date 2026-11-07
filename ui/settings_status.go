package ui

import (
	"strconv"

	"github.com/xmhrnx/giode"

	"easysend/internal/engine"
)

func (s *settingsScreen) upnpStatus() string {
	st := s.u.eng.Public()
	switch st.State {
	case engine.PublicReady:
		return "Working: reachable at " + st.IP + ":" + strconv.Itoa(st.Port) + " (UPnP)."
	case engine.PublicWorking:
		return "Asking your router…"
	case engine.PublicUnavailable:
		return st.Err
	}
	return "Asks your router (UPnP) to open a port so links work from anywhere."
}

func version() string {
	if giode.AppVersion != "" {
		return giode.AppVersion
	}
	return "1.0.0"
}
