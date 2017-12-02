package dao

import (
	"fmt"
	"github.com/revel/revel"
	"schneidernet/smarthome/app/models"
)

func TransformDeviceToDiscovery(device *Device) alexa.Endpoint {
	ep := alexa.NewEndpoint()

	ep.Capabilities = []alexa.Capability{}
	capabilities, displayCategories := GetAlexaInterfaceMapping(device.DeviceType)
	for _, deviceCap := range capabilities {
		cap := alexa.NewCapability()
		cap.Interface = deviceCap.String()
		cap.Properties.ProactivelyReported = false
		cap.Properties.Retrievable = true
		props := GetAlexaInterfaceProperties(deviceCap)
		aprops := []alexa.Property{}
		for _, p := range props {
			aprops = append(aprops, alexa.Property{Name: p})
		}
		cap.Properties.Supported = aprops
		ep.Capabilities = append(ep.Capabilities, cap)
	}

	ep.DisplayCategories = displayCategories
	ep.EndpointID = fmt.Sprintf("device-%03d", device.ID)
	ep.FriendlyName = device.Name
	ep.Description = device.Description
	ep.ManufacturerName = device.Producer

	return ep
}

func GetAlexaInterfaceProperties(cap alexa.CapabilityInterface) []string {
	retval := []string{}
	prop := func(n string) {
		retval = append(retval, n)
	}

	switch cap {
	case alexa.TemperatureSensor:
		prop("temperature")
	case alexa.BrightnessController:
		prop("brightness")
	case alexa.ColorController:
		prop("color")
	case alexa.ColorTemperaturController:
		prop("colorTemperatureInKelvin")
	case alexa.EndpointHealth:
		prop("connectivity")
	case alexa.InputController:
		prop("input")
	case alexa.LockController:
		prop("lockState")
	case alexa.PercentageController:
		prop("percentage")
	case alexa.PowerController:
		prop("powerState")
	case alexa.PowerLevelController:
		prop("powerLevel")
	case alexa.Speaker:
		prop("volume")
		prop("muted")
	case alexa.ThermostatController:
		prop("targetSetpoint")
		prop("lowerSetpoint")
		prop("upperSetpoint")
		prop("thermostatMode")
	}

	return retval
}

func GetAlexaInterfaceMapping(devicetype int) ([]alexa.CapabilityInterface, []string) {

	ifaces := []alexa.CapabilityInterface{}
	dcategories := []string{}

	iface := func(i alexa.CapabilityInterface) {
		ifaces = append(ifaces, i)
	}

	display := func(dc alexa.DisplayCategory) {
		dcategories = append(dcategories, dc.String())
	}

	// Generic Interfaces of every Capability
	iface(alexa.Alexa)
	iface(alexa.EndpointHealth)

	// Mapping
	switch devicetype {
	case alexa.DeviceSwitch.ID():
		iface(alexa.PowerController)
	case alexa.DeviceSocket.ID():
		iface(alexa.PowerController)
		display(alexa.SMARTPLUG)
	case alexa.DeviceTemperatureSensor.ID():
		iface(alexa.TemperatureSensor)
		display(alexa.TEMPERATURESENSOR)
	case alexa.DeviceLight.ID():
		iface(alexa.PowerController)
		display(alexa.LIGHT)
	case alexa.DeviceDimmableLight.ID():
		iface(alexa.BrightnessController)
		display(alexa.LIGHT)
	case alexa.DeviceColorLight.ID():
		iface(alexa.ColorController)
		iface(alexa.ColorTemperaturController)
		display(alexa.LIGHT)
	case alexa.DeviceSmartLock.ID():
		iface(alexa.LockController)
		display(alexa.DOOR)
	}
	revel.AppLog.Infof("resolved from %d types: %+v  \n %+v", devicetype, ifaces, dcategories)
	return ifaces, dcategories

}
