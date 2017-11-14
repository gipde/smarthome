package controllers

import (
	"github.com/revel/revel"
)

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	greeting := "Aloha World"
	return c.Render(greeting)
}

func (c App) Hello(myName string) revel.Result {
	c.Validation.Required(myName).Message("Your name is required!")
	c.Validation.MinSize(myName, 3).Message("Your name is not long enough!")

	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(App.Index)
	}

	return c.Render(myName)
}

func (c App) Discover() revel.Result {

	var jsonBlobS = `
	{
		"event": {
		  "header": {
			"namespace": "Alexa.Discovery",
			"name": "Discover.Response",
			"payloadVersion": "3",
			"messageId": "5f8a426e-01e4-4cc9-8b79-65f8bd0fd8a4"
		  },
		  "payload": {
			"endpoints": [
			  {
				"endpointId": "appliance-001",
				"friendlyName": "Living Room Light",
				"description": "Smart Light by Sample Manufacturer",
				"manufacturerName": "Sample Manufacturer",
				"displayCategories": [
				  "LIGHT"
				],
				"cookie": {
				  "extraDetail1": "optionalDetailForSkillAdapterToReferenceThisDevice",
				  "extraDetail2": "There can be multiple entries",
				  "extraDetail3": "but they should only be used for reference purposes",
				  "extraDetail4": "This is not a suitable place to maintain current device state"
				},
				"capabilities": [
				  {
					"type": "AlexaInterface",
					"interface": "Alexa.ColorTemperatureController",
					"version": "3",
					"properties": {
					  "supported": [
						{
						  "name": "colorTemperatureInKelvin"
						}
					  ],
					  "proactivelyReported": true,
					  "retrievable": true
					}
				  },
				  {
					"type": "AlexaInterface",
					"interface": "Alexa.EndpointHealth",
					"version": "3",
					"properties": {
					  "supported": [
						{
						  "name": "connectivity"
						}
					  ],
					  "proactivelyReported": true,
					  "retrievable": true
					}
				  },
				  {
					"type": "AlexaInterface",
					"interface": "Alexa",
					"version": "3"
				  },
				  {
					"type": "AlexaInterface",
					"interface": "Alexa.ColorController",
					"version": "3",
					"properties": {
					  "supported": [
						{
						  "name": "color"
						}
					  ],
					  "proactivelyReported": true,
					  "retrievable": true
					}
				  },
				  {
					"type": "AlexaInterface",
					"interface": "Alexa.PowerController",
					"version": "3",
					"properties": {
					  "supported": [
						{
						  "name": "powerState"
						}
					  ],
					  "proactivelyReported": true,
					  "retrievable": true
					}
				  },
				  {
					"type": "AlexaInterface",
					"interface": "Alexa.BrightnessController",
					"version": "3",
					"properties": {
					  "supported": [
						{
						  "name": "brightness"
						}
					  ],
					  "proactivelyReported": true,
					  "retrievable": true
					}
				  }
				]
			  },
			  {
				"endpointId": "appliance-002",
				"friendlyName": "Hallway Thermostat",
				"description": "Smart Thermostat by Sample Manufacturer",
				"manufacturerName": "Sample Manufacturer",
				"displayCategories": [
				  "THERMOSTAT"
				],
				"cookie": {},
				"capabilities": [
				  {
					"type": "AlexaInterface",
					"interface": "Alexa",
					"version": "3"
				  },
				  {
					"type": "AlexaInterface",
					"interface": "Alexa.ThermostatController",
					"version": "3",
					"properties": {
					  "supported": [
						{
						  "name": "lowerSetpoint"
						},
						{
						  "name": "targetSetpoint"
						},
						{
						  "name": "upperSetpoint"
						},
						{
						  "name": "thermostatMode"
						}
					  ],
					  "proactivelyReported": true,
					  "retrievable": true
					}
				  },
				  {
					"type": "AlexaInterface",
					"interface": "Alexa.TemperatureSensor",
					"version": "3",
					"properties": {
					  "supported": [
						{
						  "name": "temperature"
						}
					  ],
					  "proactivelyReported": false,
					  "retrievable": true
					}
				  }
				]
			  },
			  {
				"endpointId": "appliance-003",
				"friendlyName": "Front Door",
				"description": "Smart Lock by Sample Manufacturer",
				"manufacturerName": "Sample Manufacturer",
				"displayCategories": [
				  "SMARTLOCK"
				],
				"cookie": {},
				"capabilities": [
				  {
					"type": "AlexaInterface",
					"interface": "Alexa.LockController",
					"version": "3",
					"properties": {
					  "supported": [
						{
						  "name": "lockState"
						}
					  ],
					  "proactivelyReported": true,
					  "retrievable": true
					}
				  },
				  {
					"type": "AlexaInterface",
					"interface": "Alexa.EndpointHealth",
					"version": "3",
					"properties": {
						"supported": [
							{
								"name": "connectivity"
							}
						],
						"proactivelyReported": true,
						"retrievable": true
					}
				  }
				]
			  },
			  {
				"endpointId": "appliance-004",
				"friendlyName": "Goodnight",
				"description": "Smart Scene by Sample Manufacturer",
				"manufacturerName": "Sample Manufacturer",
				"displayCategories": [
				  "SCENE_TRIGGER"
				],
				"cookie": {},
				"capabilities": [
				  {
					"type": "AlexaInterface",
					"interface": "Alexa.SceneController",
					"version": "3",
					"supportsDeactivation": false,
					"proactivelyReported": true
				  }
				]
			  },
			  {
				"endpointId": "appliance-005",
				"friendlyName": "Watch TV",
				"description": "Smart Activity by Sample Manufacturer",
				"manufacturerName": "Sample Manufacturer",
				"displayCategories": [
				  "ACTIVITY_TRIGGER"
				],
				"cookie": {},
				"capabilities": [
				  {
					"type": "AlexaInterface",
					"interface": "Alexa",
					"version": "3"
				  },
				  {
					"type": "AlexaInterface",
					"interface": "Alexa.SceneController",
					"version": "3",
					"supportsDeactivation": true,
					"proactivelyReported": true
				  },
				  {
					"type": "AlexaInterface",
					"interface": "Alexa.EndpointHealth",
					"version": "3",
					"properties": {
						"supported": [
							{
								"name": "connectivity"
							}
						],
						"proactivelyReported": true,
						"retrievable": true
					}
				  }
				]
			  },
			  {
				"endpointId": "appliance-006",
				"friendlyName": "Back Door Camera",
				"description": "Smart Camera by Sample Manufacturer",
				"manufacturerName": "Sample Manufacturer",
				"displayCategories": [
				  "CAMERA"
				],
				"cookie": {},
				"capabilities": [
				  {
					"type": "AlexaInterface",
					"interface": "Alexa",
					"version": "3"
				  },
				  {
					"type": "AlexaInterface",
					"interface": "Alexa.CameraStreamController",
					"version": "3",
					"cameraStreamConfigurations": [
					  {
						"protocols": [
						  "RTSP"
						],
						"resolutions": [
						  {
							"width": 1920,
							"height": 1080
						  },
						  {
							"width": 1280,
							"height": 720
						  }
						],
						"authorizationTypes": [
						  "BASIC"
						],
						"videoCodecs": [
						  "H264",
						  "MPEG2"
						],
						"audioCodecs": [
						  "G711"
						]
					  },
					  {
						"protocols": [
						  "RTSP"
						],
						"resolutions": [
						  {
							"width": 1920,
							"height": 1080
						  },
						  {
							"width": 1280,
							"height": 720
						  }
						],
						"authorizationTypes": [
						  "NONE"
						],
						"videoCodecs": [
						  "H264"
						],
						"audioCodecs": [
						  "AAC"
						]
					  }
					]
				  },
				  {
					"type": "AlexaInterface",
					"interface": "Alexa.PowerController",
					"version": "3",
					"properties": {
					  "supported": [
						{
						  "name": "powerState"
						}
					  ],
					  "proactivelyReported": true,
					  "retrievable": true
					}
				  },
				  {
					"type": "AlexaInterface",
					"interface": "Alexa.EndpointHealth",
					"version": "3",
					"properties": {
						"supported": [
							{
								"name": "connectivity"
							}
						],
						"proactivelyReported": true,
						"retrievable": true
					}
				  }
				]
			  }
			]
		  }
		}
	  }
	  
	`
	c.Response.ContentType = "application/json"
	return c.RenderText(jsonBlobS)
}
