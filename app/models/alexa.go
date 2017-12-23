package alexa

import (
	"strconv"
)

type DisplayCategory int

const (
	ACTIVITYTRIGGER DisplayCategory = iota
	CAMERA
	DOOR
	LIGHT
	OTHER
	SCENETRIGGER
	SMARTLOCK
	SMARTPLUG
	SPEAKER
	SWITCH
	TEMPERATURESENSOR
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
	"SCENE_TRIGGER",
	"SMARTLOCK",
	"SMARTPLUG",
	"SPEAKER",
	"SWITCH",
	"TEMPERATURE_SENSOR",
	"THERMOSTAT",
	"TV",
}

type CapabilityInterface int

const (
	Alexa         CapabilityInterface = iota //state Reporting, implicit included
	Authorization                            // needed for Alexa event gateway
	Discovery                                // do discovery
	BrightnessController
	Calendar
	CameraStreamController
	ChannelController
	ColorController
	ColorTemperaturController
	EndpointHealth
	ErrorResponse
	InputController
	LockController
	MeetingClientController
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
	"Alexa.Calendar",
	"Alexa.CameraStreamController",
	"Alexa.ChannelController",
	"Alexa.ColorController",
	"Alexa.ColorTemperaturController",
	"Alexa.EndpointHealth",
	"Alexa.ErrorResponse",
	"Alexa.InputController",
	"Alexa.LockController",
	"Alexa.MeetingClientController",
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

type DeviceType int

const (
	DeviceSocket DeviceType = iota
	DeviceSwitch
	DeviceLight
	DeviceTemperatureSensor
	DeviceDimmableLight
	DeviceColorLight
	DeviceSmartLock

	DeviceTypeMessagePrefix = "alexa.devicetype"
	DeviceTypeNum           = len(deviceType)
)

var deviceType = [...]string{
	"Socket",
	"Switch",
	"Light",
	"TemperatureSensor",
	"DimmableLight",
	"ColorLight",
	"SmartLock",
}

func GetDeviceTypeNames() []string {
	return deviceType[:]
}

func (c DeviceType) String() string {
	return deviceType[int(c)]
}

func (c DeviceType) ID() int {
	return int(c)
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

func (c DisplayCategory) String() string {
	return displayCategories[c]
}
func (c DisplayCategory) ID() string {
	return CategoryMessageID + "." + strconv.Itoa(int(c))
}

type DiscoveryJSON struct {
	Event struct {
		Header struct {
			Namespace      string `json:"namespace"`
			Name           string `json:"name"`
			PayLoadVersion string `json:"payloadVersion"`
			MessageID      string `json:"messageId"`
		} `json:"header"`
		PayLoad struct {
			Endpoints []Endpoint `json:"endpoints"`
		} `json:"payload"`
	} `json:"event"`
}

type Endpoint struct {
	EndpointID        string        `json:"endpointId"`
	FriendlyName      string        `json:"friendlyName"`
	Description       string        `json:"description"`
	ManufacturerName  string        `json:"manufacturerName"`
	DisplayCategories []string      `json:"displayCategories"`
	Capabilities      []Capability  `json:"capabilities"`
	Cookie            []CookieEntry `json:"cookie,omitempty"`
}

type CookieEntry struct {
	Name  string
	Value string
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

// nur geschachtelte Properties
type PropBrightness struct {
	Value string `json:"value"`
}

type PropChannel struct {
	Value struct {
		Number            string `json:"number"`
		CallSign          string `json:"callSign"`
		AffiliateCallSign string `json:"affiliateCallSign"`
	} `json:"value"`
}

type PropColor struct {
	Value struct {
		Hue        float32 `json:"hue"`
		Saturation float32 `json:"saturation"`
		Brightness float32 `json:"brightness"`
	} `json:"value"`
}

type PropColorTemperatureInKelvin struct {
	Value int `json:"value"`
}

type PropConnectivity struct {
	Value struct {
		Value string `json:"value"`
	} `json:"value"`
}

type PropInput struct {
	Value string `json:"value"`
}

type PropLockSstate struct {
	Value string `json:"value"`
}

type PropMuteState struct {
	Value bool `json:"value"`
}

type PropPercentage struct {
	Value int `json:"value"`
}

// type PropPowerState struct {
// 	Value string `json:"value"`
// }

type PropTemperature struct {
	Value struct {
		Value float32 `json:"value"`
		Scale string  `json:"scale"`
	} `json:"value"`
}

type PropThermostatMode struct {
	Value      string `json:"value"`
	CustomName string `json:"customName,omitempty"`
}

type PropVolumeLevel struct {
	Volume string `json:"volume"`
}

type PropEndpointHealth struct {
	Value string `json:"value"`
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
