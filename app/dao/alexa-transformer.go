package dao

import (
	"fmt"
	"schneidernet/smarthome/app/models"
)

func TransformDeviceToDescovery(device *Device) alexa.Endpoint {
	var caps []alexa.Capability
	for _, deviceCap := range device.AlexaInterfaces {
		cap := alexa.NewCapability()
		cap.Interface = alexa.CapabilityName(deviceCap.IFace)
		cap.Properties.ProactivelyReported = false
		cap.Properties.Retrievable = true
		addApropriateProperties(&cap)

		caps = append(caps, cap)
	}
	cats := []string{}
	for _, deviceCat := range device.DisplayCategories {
		cats = append(cats, alexa.DisplayCategoryName(deviceCat.DCat))

	}
	ep := alexa.NewEndpoint()
	ep.EndpointID = fmt.Sprintf("device-%03d", device.ID)
	ep.FriendlyName = device.Name
	ep.Description = device.Description
	ep.ManufacturerName = device.Producer
	ep.Capabilities = caps
	ep.DisplayCategories = cats

	return ep
}

func addApropriateProperties(cap *alexa.Capability) {
	switch cap.Interface {
	case alexa.TemperatureSensor.String():
		cap.Properties.Supported = append(cap.Properties.Supported, alexa.Property{Name: "temperature"})
	}
}
