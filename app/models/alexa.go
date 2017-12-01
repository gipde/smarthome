package alexa

import (
	"strconv"
)

type DisplayCategory int

const (
	ACTIVITY_TRIGGER DisplayCategory = iota
	CAMERA
	DOOR
	LIGHT
	OTHER
	SCENETRIGGER
	SMARTLOCK
	SMARTPLUG
	SPEAKERS
	SWITCH
	TEMPERATURSENSOR
	THERMOSTAT
	TV

	CategoryMessageID = "alexa.display_category"
	CategoryNum       = len(displayCategories)
)

var displayCategories = [...]string{
	"ACTIVITY_TRIGGER",
	"CAMERA",
	"DOOR",
	"LIGHT",
	"OTHER",
	"SCENETRIGGER",
	"SMARTLOCK",
	"SMARTPLUG",
	"SPEAKERS",
	"SWITCH",
	"TEMPERATURSENSOR",
	"THERMOSTAT",
	"TV",
}

type CapabilityInterface int

const (
	Alexa CapabilityInterface = iota
	Authorization
	Discovery
	BrightnessController
	CameraStreamController
	ChannelController
	ColorController
	ColorTemperaturController
	EndpointHealth
	ErrorResponse
	InputController
	LockController
	PercentageController
	PlaybackController
	PowerController
	PowerLevelController
	SceneController
	Speaker
	StepSpeaker
	TemperatureSensor
	ThermostatController

	CapabilityInterfaceMessageID = "alexa.cap_interface"
	CapabilityInterfaceNums      = len(capabilityInterfaces)
)

var capabilityInterfaces = [...]string{
	"Alexa",
	"Alexa.Authorization",
	"Alexa.Discovery",
	"Alexa.BrightnessController",
	"Alexa.CameraStreamController",
	"Alexa.ChannelController",
	"Alexa.ColorController",
	"Alexa.ColorTemperaturController",
	"Alexa.EndpointHealth",
	"Alexa.ErrorResponse",
	"Alexa.InputController",
	"Alexa.LockController",
	"Alexa.PercentageController",
	"Alexa.PlaybackController",
	"Alexa.PowerController",
	"Alexa.PowerLevelController",
	"Alexa.SceneController",
	"Alexa.Speaker",
	"Alexa.StepSpeaker",
	"Alexa.TemperatureSensor",
	"Alexa.ThermostatController",
}

func (c CapabilityInterface) String() string {
	return capabilityInterfaces[int(c)]
}
func CapabilityName(num int) string {
	return capabilityInterfaces[num]
}

func DisplayCategoryName(num int) string {
	return displayCategories[num]
}

func (c DisplayCategory) ID() string {
	return CategoryMessageID + "." + strconv.Itoa(int(c))
}

type DiscoveryJSON struct {
	Event struct {
		Header struct {
			Namespace      string `json:"namespace"`
			Name           string `json:"version"`
			PayLoadVersion string `json:"payloadVersion"`
			MessageID      string `json:"messageId"`
		} `json:"header"`
		PayLoad struct {
			Endpoints []Endpoint `json:"endpoints"`
		} `json:"payload"`
	} `json:"event"`
}

type Endpoint struct {
	EndpointID        string       `json:"endpointId"`
	FriendlyName      string       `json:"friendlyName"`
	Description       string       `json:"description"`
	ManufacturerName  string       `json:"manufacturerName"`
	DisplayCategories []string     `json:"displayCategories"`
	Capabilities      []Capability `json:"capabilities"`
}

type Capability struct {
	Type       string `json:"type"`
	Interface  string `json:"interface"`
	Version    string `json:"version"`
	Properties struct {
		Supported           []Property `json:"supported,omitempty"`
		ProactivelyReported bool       `json:"proactivelyReported"`
		Retrievable         bool       `json:"retrievable"`
	} `json:"properties,omitempty"`
}

type Property struct {
	Name string `json:"name"`
}

type Brightness struct {
	Value string `json:"value"`
}

type Channel struct {
	Value struct {
		Number            string `json:"number"`
		CallSign          string `json:"callSign"`
		AffiliateCallSign string `json:"affiliateCallSign"`
	} `json:"value"`
}

type Color struct {
	Value struct {
		Hue        float32 `json:"hue"`
		Saturation float32 `json:"saturation"`
		Brightness float32 `json:"brightness"`
	} `json:"value"`
}

type ColorTemperatureInKelvin struct {
	Value int `json:"value"`
}

type Connectivity struct {
	Value struct {
		Value string `json:"value"`
	} `json:"value"`
}

type Input struct {
	Value string `json:"value"`
}

type LockSstate struct {
	Value string `json:"value"`
}

type MuteState struct {
	Value bool `json:"value"`
}

type Percentage struct {
	Value int `json:"value"`
}

type PowerState struct {
	Value string `json:"value"`
}

type Temperature struct {
	Value struct {
		Value float32 `json:"value"`
		Scale string  `json:"scale"`
	} `json:"value"`
}

type ThermostatMode struct {
	Value      string `json:"value"`
	CustomName string `json:"customName,omitempty"`
}

type VolumeLevel struct {
	Volume string `json:"volume"`
}

func NewProperties(vals []string) []Property {
	r := []Property{}
	for _, v := range vals {
		r = append(r, Property{Name: v})
	}
	return r
}

func NewDiscovery(uuid string) DiscoveryJSON {
	resp := DiscoveryJSON{}
	resp.Event.Header.MessageID = uuid
	resp.Event.Header.Namespace = "Alexa.Discovery"
	resp.Event.Header.Name = "Discover.Response"
	resp.Event.Header.PayLoadVersion = "3"
	return resp
}

func NewCapability() Capability {
	r := Capability{}
	r.Type = "AlexaInterface"
	r.Version = "3"
	return r
}

func NewEndpoint() Endpoint {
	return Endpoint{}
}
