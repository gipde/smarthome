// Package alexa provides the primary model for interacting with alexa
package alexa

import (
	"strconv"
)

// DisplayCategories
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

const (
	VERSION3 = "3"

	ON  = "ON"
	OFF = "OFF"

	OK          = "OK"
	UNREACHABLE = "UNREACHABLE"
)

// CapabilityInterface Core Interface Types
const (
	Alexa CapabilityInterface = iota
	Authorization
	Discovery
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

// DeviceTypes
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

// Core JSON Types

// Request primary Request Structure
type Request struct {
	Directive struct {
		Header   Header
		Endpoint Endpoint
		Payload  struct {
			Scope Scope
		}
	}
}

// DiscoveryResponse for Discovery Requests
type DiscoveryResponse struct {
	Event struct {
		Header struct {
			Namespace      string `json:"namespace"`
			Name           string `json:"name"`
			PayLoadVersion string `json:"payloadVersion"`
			MessageID      string `json:"messageId"`
		} `json:"header"`
		PayLoad struct {
			DiscoveryEndpoints []DiscoveryEndpoint `json:"endpoints"`
		} `json:"payload"`
	} `json:"event"`
}

// StateReport Response Structure
type StateReport struct {
	Context struct {
		Properties []Properties `json:"properties"`
	} `json:"context"`
	Event struct {
		Header   Header   `json:"header"`
		Endpoint Endpoint `json:"endpoint"`
		PayLoad  struct{} `json:"payload"`
	} `json:"event"`
}

// Sub-JSON Fragments

// Endpoint JSON Fragment for DiscoveryResponse
type DiscoveryEndpoint struct {
	EndpointID        string        `json:"endpointId"`
	FriendlyName      string        `json:"friendlyName"`
	Description       string        `json:"description"`
	ManufacturerName  string        `json:"manufacturerName"`
	DisplayCategories []string      `json:"displayCategories"`
	Capabilities      []Capability  `json:"capabilities"`
	Cookie            []CookieEntry `json:"cookie,omitempty"`
}

// Capability Fragment in Endpoint struct
type Capability struct {
	Type       string             `json:"type"`
	Interface  string             `json:"interface"`
	Version    string             `json:"version"`
	Properties CapabilityProperty `json:"properties,omitempty"`
}

type CapabilityProperty struct {
	Supported           []Property `json:"supported,omitempty"`
	ProactivelyReported bool       `json:"proactivelyReported"`
	Retrievable         bool       `json:"retrievable"`
}

// CookieEntry Fragment in Endpoint struct
type CookieEntry struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Header JSON Fragment
type Header struct {
	Namespace        string `json:"namespace"`
	Name             string `json:"name"`
	PayloadVersion   string `json:"payloadVersion"`
	MessageID        string `json:"messageId"`
	CorrelationToken string `json:"correlationToken"`
}

// Endpoint JSON Fragment
type Endpoint struct {
	EndpointID string `json:"endpointId"`
	Scope      Scope  `json:"scope"`
}

// Scope JSON Fragment
type Scope struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

// Properties JSON Fragment
type Properties struct {
	Namesapce                 string      `json:"namespace"`
	Name                      string      `json:"name"`
	Value                     interface{} `json:"value"`
	TimeOfSample              string      `json:"timeOfSample"`
	UncertaintyInMilliseconds int         `json:"uncertaintyInMilliseconds"`
}

// Property JSON Fragment
type Property struct {
	Name string `json:"name"`
}

// Alexa Property Types

// PropBrightness Brightness Setting
type PropBrightness struct {
	Value string `json:"value"`
}

// PropChannel for Calling
type PropChannel struct {
	Value struct {
		Number            string `json:"number"`
		CallSign          string `json:"callSign"`
		AffiliateCallSign string `json:"affiliateCallSign"`
	} `json:"value"`
}

// PropColor for Color Lights
type PropColor struct {
	Value struct {
		Hue        float32 `json:"hue"`
		Saturation float32 `json:"saturation"`
		Brightness float32 `json:"brightness"`
	} `json:"value"`
}

// PropColorTemperatureInKelvin for Kelvin ColorTemperature
type PropColorTemperatureInKelvin struct {
	Value int `json:"value"`
}

//PropConnectivity mainly used for Endpoint Health
type PropConnectivity struct {
	Value struct {
		Value string `json:"value"`
	} `json:"value"`
}

// PropInput Input Value
type PropInput struct {
	Value string `json:"value"`
}

// PropLockState used for Doors
type PropLockState struct {
	Value string `json:"value"`
}

// PropMuteState ...
type PropMuteState struct {
	Value bool `json:"value"`
}

// PropPercentage for Dimmable Lights
type PropPercentage struct {
	Value int `json:"value"`
}

// PropTemperature for TemperatureSensors
type PropTemperature struct {
	Value struct {
		Value float32 `json:"value"`
		Scale string  `json:"scale"`
	} `json:"value"`
}

// PropThermostatMode for Thermostats
type PropThermostatMode struct {
	Value      string `json:"value"`
	CustomName string `json:"customName,omitempty"`
}

// PropVolumeLevel for Speakers
type PropVolumeLevel struct {
	Volume string `json:"volume"`
}

// PropEndpointHealth ...
type PropEndpointHealth struct {
	Value string `json:"value"`
}

// ******************************************

// DeviceType categorizes all Alexa Device-Types
type DeviceType int

// String representation of DeviceType
func (c DeviceType) String() string {
	return deviceType[int(c)]
}

// ID of DeviceType
func (c DeviceType) ID() int {
	return int(c)
}

// GetDeviceTypeNames gives all device-type Names as a slice
func GetDeviceTypeNames() []string {
	return deviceType[:]
}

// ******************************************

// CapabilityInterface is a Type for alle Alexa Interfaces
type CapabilityInterface int

// Stringer interface
func (c CapabilityInterface) String() string {
	return capabilityInterfaces[int(c)]
}

// CapabilityName -> String representation
func CapabilityName(num int) string {
	return capabilityInterfaces[num]
}

// ******************************************

// DisplayCategory is a type for Display Icons in the Alexa-App
// https://developer.amazon.com/de/docs/device-apis/alexa-discovery.html#display-categories
type DisplayCategory int

// Stringer interface
func (c DisplayCategory) String() string {
	return displayCategories[c]
}

// ID for message-file
func (c DisplayCategory) ID() string {
	return CategoryMessageID + "." + strconv.Itoa(int(c))
}

// DisplayCategoryName String representation
func DisplayCategoryName(num int) string {
	return displayCategories[num]
}

// NewProperties from a string-array
func NewProperties(vals []string) []Property {
	r := []Property{}
	for _, v := range vals {
		r = append(r, Property{Name: v})
	}
	return r
}

// ******************************************

// NewDiscoveryResponse with a uuid
func NewDiscoveryResponse(uuid string) DiscoveryResponse {
	resp := DiscoveryResponse{}
	resp.Event.Header.MessageID = uuid
	resp.Event.Header.Namespace = Discovery.String()
	resp.Event.Header.Name = "Discover.Response"
	resp.Event.Header.PayLoadVersion = VERSION3
	return resp
}

// NewCapability ...
func NewCapability() Capability {
	r := Capability{}
	r.Type = "AlexaInterface"
	r.Version = VERSION3
	return r
}
