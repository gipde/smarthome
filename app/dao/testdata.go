package dao

import (
	"schneidernet/smarthome/app/models/alexa"
)

func GetTestDevices() *[]Device {

	return &[]Device{
		Device{
			Name:        "Licht Küche",
			Description: "Licht Küche unter der Theke",
			Producer:    "Schneidernet",
			DeviceType:  alexa.DeviceLight.ID(),
			State:       alexa.ON,
			Connected:   true,
		},
		Device{
			Name:        "Warmwasser Heizung",
			Description: "Warmwasser Temperatur der Heizung",
			Producer:    "Schneidernet",
			DeviceType:  alexa.DeviceTemperatureSensor.ID(),
			State:       "35.5",
			Connected:   false,
		},
		Device{
			Name:        "TV",
			Description: "Heizung",
			Producer:    "Schneidernet",
			DeviceType:  alexa.DeviceSwitch.ID(),
			State:       alexa.ON,
			Connected:   false,
		},
		Device{
			Name:        "Steckdose",
			Description: "Esszimmer",
			Producer:    "Schneidernet",
			DeviceType:  alexa.DeviceSocket.ID(),
			State:       alexa.OFF,
			Connected:   true,
		},
	}
}
