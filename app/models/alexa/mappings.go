package alexa

// GetAlexaInterfaceProperties gets all interface properties
func GetAlexaInterfaceProperties(cap CapabilityInterface) []string {

	retval := []string{}
	prop := func(n string) {
		retval = append(retval, n)
	}

	switch cap {
	case TemperatureSensor:
		prop("temperature")
	case BrightnessController:
		prop("brightness")
	case ColorController:
		prop("color")
	case ColorTemperaturController:
		prop("colorTemperatureInKelvin")
	case EndpointHealth:
		prop("connectivity")
	case InputController:
		prop("input")
	case LockController:
		prop("lockState")
	case PercentageController:
		prop("percentage")
	case PowerController:
		prop("powerState")
	case PowerLevelController:
		prop("powerLevel")
	case Speaker:
		prop("volume")
		prop("muted")
	case ThermostatController:
		prop("targetSetpoint")
		prop("lowerSetpoint")
		prop("upperSetpoint")
		prop("thermostatMode")
	}

	return retval
}

//GetAlexaInterfaceMapping gets mapping for eatch devicetype
func GetAlexaInterfaceMapping(devicetype int) ([]CapabilityInterface, []string) {

	ifaces := []CapabilityInterface{}
	dcategories := []string{}

	iface := func(i CapabilityInterface) {
		ifaces = append(ifaces, i)
	}

	display := func(dc DisplayCategory) {
		dcategories = append(dcategories, dc.String())
	}

	// Generic Interfaces of every Capability
	iface(Alexa)
	iface(EndpointHealth)

	// Mapping
	switch devicetype {
	case DeviceSwitch.ID():
		iface(PowerController)
		display(SWITCH)
	case DeviceSocket.ID():
		iface(PowerController)
		display(SMARTPLUG)
	case DeviceTemperatureSensor.ID():
		iface(TemperatureSensor)
		display(TEMPERATURESENSOR)
	case DeviceLight.ID():
		iface(PowerController)
		display(LIGHT)
	case DeviceDimmableLight.ID():
		iface(BrightnessController)
		display(LIGHT)
	case DeviceColorLight.ID():
		iface(ColorController)
		iface(ColorTemperaturController)
		display(LIGHT)
	case DeviceSmartLock.ID():
		iface(LockController)
		display(DOOR)
	}

	return ifaces, dcategories
}
