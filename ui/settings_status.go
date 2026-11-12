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
		return tr("Working: reachable at {address} (UPnP).", "address", st.IP+":"+strconv.Itoa(st.Port))
	case engine.PublicWorking:
		return tr("Asking your router…")
	case engine.PublicUnavailable:
		return tr(st.Err)
	}
	return tr("Asks your router (UPnP) to open a port so links work from anywhere.")
}

func version() string {
	if giode.AppVersion != "" {
		return giode.AppVersion
	}
	return "1.0.0"
}
