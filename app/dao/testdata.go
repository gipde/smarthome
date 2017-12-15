package dao

import (
	"schneidernet/smarthome/app/models"
)

func GetTestDevices() *[]Device {

	return &[]Device{
		Device{
			Name:        "Licht Küche",
			Description: "Licht Küche unter der Theke",
			Producer:    "Schneidernet",
			DeviceType:  alexa.DeviceLight.ID(),
			State:       "ON",
			Connected:   true,
		},
		Device{
			Name:        "Temperatur Sensor",
			Description: "Heizung",
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
			State:       "ON",
			Connected:   false,
		},
		Device{
			Name:        "Steckdose",
			Description: "Esszimmer",
			Producer:    "Schneidernet",
			DeviceType:  alexa.DeviceSocket.ID(),
			State:       "OFF",
			Connected:   true,
		},
	}
}
