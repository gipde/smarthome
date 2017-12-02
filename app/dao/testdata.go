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
		},
		Device{
			Name:        "Temperatur Sensor",
			Description: "Esszimmer",
			Producer:    "Schneidernet",
			DeviceType:  alexa.DeviceTemperatureSensor.ID(),
		},
	}
}
