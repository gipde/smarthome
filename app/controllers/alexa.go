package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/revel/revel"
	"schneidernet/smarthome/app"
	"schneidernet/smarthome/app/dao"
	"schneidernet/smarthome/app/models/alexa"
	"time"
)

const (
	alexaTimeFormat = "2006-01-02T15:04:05.00Z"
)

// Alexa Controler for Amazon Echo
type Alexa struct {
	*revel.Controller
}

// API is the main Entry Point for Alexa
func (c Alexa) API(r alexa.Request) revel.Result {

	var user *dao.User
	var err error
	if user, err = checkUser(&r); err != nil {
		return c.RenderError(err)
	}

	c.Response.ContentType = "application/json"

	// Special Case Discovery
	if r.Directive.Header.Name == "Discover" {
		return c.discovery(user.ID)
	}

	device := dao.FindDeviceByID(user.ID, r.Directive.Endpoint.EndpointID)

	// Switch on Request-Type
	switch r.Directive.Header.Name {
	case "ReportState":
		return c.reportStateHandler(&r, device, "StateReport")
	case "TurnOn":
		return c.switchHandler(&r, device, alexa.ON, "Response")
	case "TurnOff":
		return c.switchHandler(&r, device, alexa.OFF, "Response")
	}

	// Default
	c.Response.Status = 500
	return c.RenderText("Error")
}

// discovery Request
func (c Alexa) discovery(useroid uint) revel.Result {
	response := generateDiscoveryResponse(dao.GetAllDevices(useroid))
	value, _ := json.MarshalIndent(response, "", "   ")
	c.Log.Debugf("discovery %s", string(value))
	return c.RenderJSON(response)
}

// Report State / Response on Change
func (c Alexa) reportStateHandler(request *alexa.Request, device *dao.Device, headerName string) revel.Result {

	// StateReport Base-Structure
	var s alexa.StateReport
	s.Event.Header = alexa.Header{
		Namespace:        alexa.Alexa.String(),
		Name:             headerName,
		PayloadVersion:   alexa.VERSION3,
		CorrelationToken: request.Directive.Header.CorrelationToken,
		MessageID:        app.NewUUID(),
	}
	s.Event.Endpoint = alexa.Endpoint{
		EndpointID: request.Directive.Endpoint.EndpointID,
		Scope: alexa.Scope{
			Type:  request.Directive.Endpoint.Scope.Type,
			Token: request.Directive.Endpoint.Scope.Token,
		},
	}

	// Generate all Properties
	s.Context.Properties = []alexa.Properties{}

	alexaInterfaces, _ := alexa.GetAlexaInterfaceMapping(device.DeviceType)
	for _, iface := range alexaInterfaces {

		//TODO: context sensitive Properties
		for _, pname := range alexa.GetAlexaInterfaceProperties(iface) {

			// BaseData of every Property
			prop := alexa.Properties{
				Namesapce:                 iface.String(),
				Name:                      pname,
				TimeOfSample:              time.Now().Format(alexaTimeFormat),
				UncertaintyInMilliseconds: 500,
			}
			// Individual Values of each Property
			switch iface {
			case alexa.PowerController:
				prop.Value = device.State
			case alexa.EndpointHealth:
				prop.Value = alexa.PropEndpointHealth{Value: alexa.OK}
			case alexa.TemperatureSensor:
				prop.Value = "28.4" // Geschachtelt Value { value: 12.3, scale: "CELSIUS" }
			}

			s.Context.Properties = append(s.Context.Properties, prop)

		}
	}
	return c.RenderJSON(s)
}

// switch Handler
func (c Alexa) switchHandler(request *alexa.Request, device *dao.Device, state string, headerName string) revel.Result {
	device.State = state
	SetState(device.ID, state)
	return c.reportStateHandler(request, device, headerName)
}

// maps a database entry to a json equivalent discovery response
func generateDiscoveryResponse(devices *[]dao.Device) alexa.DiscoveryResponse {
	resp := alexa.NewDiscoveryResponse(app.NewUUID())

	var eps []alexa.DiscoveryEndpoint
	for _, device := range *devices {
		eps = append(eps, getDiscoveryEndpoint(&device))
	}

	resp.Event.PayLoad.DiscoveryEndpoints = eps
	return resp
}

// check User for this Request
func checkUser(r *alexa.Request) (*dao.User, error) {

	// Get Token and check User
	var token string
	if r.Directive.Header.Namespace == alexa.Discovery.String() {
		token = r.Directive.Payload.Scope.Token
	} else {
		token = r.Directive.Endpoint.Scope.Token
	}

	valid, username := CheckToken(token)

	user := dao.GetUser(username)
	if valid == false {
		return nil, errors.New("Invalid Token")
	}
	if user == nil {
		return nil, errors.New("User not found")
	}

	return user, nil
}

// getDiscoveryEndpoint from a device
func getDiscoveryEndpoint(device *dao.Device) alexa.DiscoveryEndpoint {
	capabilities, displayCategories := alexa.GetAlexaInterfaceMapping(device.DeviceType)

	ep := alexa.DiscoveryEndpoint{
		Capabilities:      []alexa.Capability{},
		DisplayCategories: displayCategories,
		EndpointID:        fmt.Sprintf("device-%d", device.ID),
		FriendlyName:      device.Name,
		Description:       device.Description,
		ManufacturerName:  device.Producer,
	}

	for _, deviceCap := range capabilities {

		cap := alexa.NewCapability()
		cap.Interface = deviceCap.String()

		props := alexa.GetAlexaInterfaceProperties(deviceCap)
		aprops := []alexa.Property{}
		for _, p := range props {
			aprops = append(aprops, alexa.Property{Name: p})
		}

		cap.Properties = alexa.CapabilityProperty{
			// TODO: optimize this
			ProactivelyReported: false,
			Retrievable:         true,
			Supported:           aprops,
		}

		ep.Capabilities = append(ep.Capabilities, cap)
	}

	return ep
}
