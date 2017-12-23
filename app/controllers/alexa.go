package controllers

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/revel/revel"
	"io"
	"io/ioutil"
	"net/http"
	"schneidernet/smarthome/app"
	"schneidernet/smarthome/app/dao"
	"schneidernet/smarthome/app/models"
	"time"
)

const rpi = "http://ds1820ws/"

type Controller struct {
	State string
}

type Request struct {
	Directive struct {
		Header struct {
			MessageID        string
			Name             string
			Namespace        string
			PayloadVersion   string
			CorrelationToken string
		}
		Endpoint struct {
			Scope struct {
				Type  string
				Token string
			}
			EndpointID string
		}
		Payload struct {
			Scope struct {
				Type  string
				Token string
			}
		}
	}
}

type Properties struct {
	Namesapce                 string      `json:"namespace"`
	Name                      string      `json:"name"`
	Value                     interface{} `json:"value"`
	TimeOfSample              string      `json:"timeOfSample"`
	UncertaintyInMilliseconds int         `json:"uncertaintyInMilliseconds"`
}
type StateReport struct {
	Context struct {
		Properties []Properties `json:"properties"`
	} `json:"context"`
	Event struct {
		Header struct {
			Namesapce        string `json:"namespace"`
			Name             string `json:"name"`
			PayloadVersion   string `json:"payloadVersion"`
			MessageID        string `json:"messageId"`
			CorrelationToken string `json:"correlationToken"`
		} `json:"header"`
		Endpoint struct {
			EndpointID string `json:"endpointId"`
			Scope      struct {
				Type  string `json:"type"`
				Token string `json:"token"`
			} `json:"scope"`
		} `json:"endpoint"`
		PayLoad struct {
		} `json:"payload"`
	} `json:"event"`
}

func (c Alexa) reportStateHandler(request Request, device *dao.Device) revel.Result {
	var s StateReport
	s.Event.Endpoint.EndpointID = request.Directive.Endpoint.EndpointID
	s.Event.Header.Namesapce = "Alexa"
	s.Event.Header.Name = "StateReport"
	s.Event.Header.PayloadVersion = "3"
	s.Event.Header.CorrelationToken = request.Directive.Header.CorrelationToken
	s.Event.Header.MessageID = newUUID()
	s.Event.Endpoint.Scope.Type = request.Directive.Endpoint.Scope.Type
	s.Event.Endpoint.Scope.Token = request.Directive.Endpoint.Scope.Token

	c.Log.Infof("Device: %+v ", *device)
	s.Context.Properties = []Properties{}
	alexaInterfaces, _ := dao.GetAlexaInterfaceMapping(device.DeviceType)
	for _, iface := range alexaInterfaces {
		c.Log.Infof("Loop Interfaces: %+v", iface)

		//TODO: context sensitive Properties
		for _, pname := range dao.GetAlexaInterfaceProperties(iface) {
			c.Log.Infof("Loop Properties: %+v", pname)
			prop := Properties{
				Namesapce:                 iface.String(),
				Name:                      pname,
				TimeOfSample:              time.Now().Format("2006-01-02T15:04:05.00Z"),
				UncertaintyInMilliseconds: 500,
			}
			switch iface {
			case alexa.PowerController:
				prop.Value = "OFF" // OFF
			case alexa.EndpointHealth:
				prop.Value = alexa.PropEndpointHealth{Value: "OK"} // UNREACHABLE
			case alexa.TemperatureSensor:
				prop.Value = "28.4" // Geschachtelt Value { value: 12.3, scale: "CELSIUS" }
			}

			s.Context.Properties = append(s.Context.Properties, prop)

		}
	}
	c.Log.Infof("Report State for device %+v answer %+v", device, s)
	return c.RenderJSON(s)
}

type Alexa struct {
	*revel.Controller
}

func (c Alexa) Api(r Request) revel.Result {
	// c.Log.Debugf("API Request: %+v", r)
	DoLogRevelRequest(c.Request, "Alexa Api Call")

	var token string
	if r.Directive.Header.Namespace == "Alexa.Discovery" {
		token = r.Directive.Payload.Scope.Token
	} else {
		token = r.Directive.Endpoint.Scope.Token
	}
	valid, username := app.CheckToken(token)
	c.Log.Infof("valid: %b user: %s", valid, username)

	user := dao.GetUser(username)
	if valid == false {
		return c.RenderError(errors.New("invalid Token"))
	}
	if user == nil {
		return c.RenderError(errors.New("User not found"))
	}

	c.Response.ContentType = "application/json"

	switch r.Directive.Header.Name {

	case "Discover":
		return c.discovery(user.ID)

	case "ReportState":
		return c.reportStateHandler(
			r, dao.FindDeviceByID(user.ID, r.Directive.Endpoint.EndpointID))
	case "TurnOn":
		return c.doSwitch("ON")
	case "TurnOff":
		return c.doSwitch("OFF")

	}
	c.Response.Status = 500
	return c.discovery(user.ID)
	//return c.RenderText("Error")
}

func (c Alexa) discovery(useroid uint) revel.Result {
	response := generateDiscoveryResponse(dao.GetAllDevicesDeep(useroid))
	return c.RenderJSON(response)
}

// maps a database entry to a json equivalent discovery response
func generateDiscoveryResponse(devices *[]dao.Device) alexa.DiscoveryJSON {
	resp := alexa.NewDiscovery(newUUID())
	var eps []alexa.Endpoint

	for _, device := range *devices {
		eps = append(eps, dao.TransformDeviceToDiscovery(&device))
	}
	resp.Event.PayLoad.Endpoints = eps
	return resp
}

func (c Alexa) doSwitch(state string) revel.Result {
	c.Log.Info("SWITCH State: " + state)
	response := doHTTP(rpi + state)
	c.ViewArgs["powerState"] = getController(response).State
	return c.RenderTemplate("Alexa/schalter.switch.json")
}

func getSwitchState() string {
	response := doHTTP(rpi + "state")
	return getController(response).State
}
func getController(jsonData []byte) Controller {
	var pc Controller
	json.Unmarshal(jsonData, &pc)
	revel.AppLog.Info("Got State: " + pc.State)
	return pc
}

func (c Alexa) prettyprint(jsonData Request) {
	st, _ := json.MarshalIndent(jsonData, "", "    ")
	c.Log.Info(string(st))
}

func newUUID() string {
	uuid := make([]byte, 16)
	io.ReadFull(rand.Reader, uuid)
	uuid[8] = uuid[8]&^0xc0 | 0x80
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

func doHTTP(uri string) []byte {
	resp, err := http.Get(uri)
	if err != nil {
		revel.AppLog.Error("Error getting Controller State: " + err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		revel.AppLog.Error("Error getting Request Body: " + err.Error())
	}
	revel.AppLog.Info("Calling: " + uri + " got " + string(body))
	return body
}
