// Package localsend implements the LocalSend v2 protocol, so EasySend
// discovers and exchanges files with other EasySend devices and with the
// LocalSend apps on the local network.
//
// Protocol: https://github.com/localsend/protocol
package localsend

import "errors"

// Protocol constants.
const (
	MulticastGroup = "224.0.0.167"
	DefaultPort    = 53317
	ProtocolVer    = "2.1"
	apiBase        = "/api/localsend/v2"
)

// Device types LocalSend knows.
const (
	TypeMobile   = "mobile"
	TypeDesktop  = "desktop"
	TypeWeb      = "web"
	TypeHeadless = "headless"
	TypeServer   = "server"
)

// Errors returned when sending.
var (
	ErrPINRequired = errors.New("the receiver requires a PIN")
	ErrDeclined    = errors.New("the receiver declined")
	ErrBusy        = errors.New("the receiver is busy with another transfer")
	ErrCanceled    = errors.New("canceled")
)
