package devcom

import ()

type DevProto struct {
	Device  `json:"device,omitempty"`
	Action  string      `json:"action"`
	PayLoad interface{} `json:"payload,omitempty"`
}
type Device struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Connected   bool   `json:"connected"`
	DeviceType  int    `json:"devicetype"`
}

const (
	Connect     = "Connect"
	Disconnect  = "Disconnect"
	FlipState   = "FlipState"
	SetState    = "SetState"
	StateUpdate = "StateUpdate"

	RequestState  = "RequestState"
	StateResponse = "StateResponse"

	Ping = "Ping"
	Pong = "Pong"

	ListeDevices = "ListDevices"
	DeviceList   = "DeviceList"
)
