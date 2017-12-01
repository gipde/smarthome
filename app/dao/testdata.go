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
			AlexaInterfaces: []AlexaInterface{
				AlexaInterface{
					IFace: int(alexa.PowerController),
				},
				AlexaInterface{IFace: int(alexa.EndpointHealth)},
			},
			DisplayCategories: []DisplayCategory{
				DisplayCategory{DCat: int(alexa.SMARTPLUG)},
			},
		},
		Device{
			Name:        "Temperatur Sensor",
			Description: "Esszimmer",
			Producer:    "Schneidernet",
			AlexaInterfaces: []AlexaInterface{
				AlexaInterface{
					IFace: int(alexa.TemperatureSensor),
				},
				AlexaInterface{IFace: int(alexa.EndpointHealth)},
			},
			DisplayCategories: []DisplayCategory{
				DisplayCategory{DCat: int(alexa.LIGHT)},
			},
		},
	}
}
