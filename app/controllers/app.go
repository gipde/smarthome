package controllers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/revel/revel"
	"io"
	"io/ioutil"
	"net/http"
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
		Payload map[string]interface{}
	}
}

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	return c.Alexa()
}

func (c App) Alexa() revel.Result {

	var r Request
	c.Params.BindJSON(&r)

	c.Response.ContentType = "application/json"

	c.ViewArgs["corrId"] = r.Directive.Header.CorrelationToken
	c.ViewArgs["uuid"] = newUUID()
	c.ViewArgs["timestamp"] = time.Now().Format("2006-01-02T15:04:05.00Z")

	switch r.Directive.Header.Name {
	case "Discover":
		return c.RenderTemplate("App/discovery.json")

	case "ReportState":
		c.Log.Info("Report State: " + r.Directive.Endpoint.EndpointID)
		switch r.Directive.Endpoint.EndpointID {
		case "heizung-001":
			return c.RenderTemplate("App/heating.state.json")
		case "zirkulationspumpe-001":
			c.ViewArgs["powerState"] = getSwitchState()
			return c.RenderTemplate("App/schalter.state.json")
		}

	case "TurnOn":
		return c.doSwitch("ON")
	case "TurnOff":
		return c.doSwitch("OFF")

	}
	c.Response.Status = 500
	return c.RenderText("Error")
}

func (c App) doSwitch(state string) revel.Result {
	c.Log.Info("SWITCH State: " + state)
	response := doHTTP(rpi + state)
	c.ViewArgs["powerState"] = getController(response).State
	return c.RenderTemplate("App/schalter.switch.json")
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

func (c App) prettyprint(jsonData Request) {
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
