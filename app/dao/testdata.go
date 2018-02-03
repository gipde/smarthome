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
			Connected:   "client",
		},
		Device{
			Name:        "Warmwasser Heizung",
			Description: "Warmwasser Temperatur der Heizung",
			Producer:    "Schneidernet",
			DeviceType:  alexa.DeviceTemperatureSensor.ID(),
			State:       "35.5",
		},
		Device{
			Name:        "TV",
			Description: "Heizung",
			Producer:    "Schneidernet",
			DeviceType:  alexa.DeviceSwitch.ID(),
			State:       alexa.ON,
		},
		Device{
			Name:        "Steckdose",
			Description: "Esszimmer",
			Producer:    "Schneidernet",
			DeviceType:  alexa.DeviceSocket.ID(),
			State:       alexa.OFF,
		},
	}
}
