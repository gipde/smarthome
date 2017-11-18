package controllers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/revel/revel"
	"io"
	"time"
)

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	greeting := "Aloha World"
	return c.Render(greeting)
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

var powerState = "ON"

func (c App) Alexa() revel.Result {

	var r Request

	c.Params.BindJSON(&r)
	c.prettyprint(r)
	c.Response.ContentType = "application/json"

	c.ViewArgs["corrId"] = r.Directive.Header.CorrelationToken
	c.ViewArgs["uuid"] = newUUID()
	c.ViewArgs["timestamp"] = time.Now().Format("2006-01-02T15:04:05.00Z")

	switch r.Directive.Header.Name {
	case "Discover":
		return c.RenderTemplate("App/discovery.json")

	case "ReportState":
		switch r.Directive.Endpoint.EndpointID {
		case "heizung-001":
			return c.RenderTemplate("App/heating.state.json")
		case "schalter-001":
			c.ViewArgs["powerState"] = powerState
			return c.RenderTemplate("App/schalter.state.json")
		}

	case "TurnOn":
		return c.doSwitch("ON")
	case "TurnOff":
		return c.doSwitch("OFF")

	}

	return c.RenderText("Hello")
}

func (c App) doSwitch(state string) revel.Result {
	c.Log.Info("SWITCH State: " + state)
	powerState = state
	c.ViewArgs["powerState"] = powerState
	return c.RenderTemplate("App/schalter.toggle.json")
}
func (c App) prettyprint(jsonData Request) {
	st, _ := json.MarshalIndent(jsonData, "", "    ")
	c.Log.Info(string(st))
}

func newUUID() string {
	uuid := make([]byte, 16)
	io.ReadFull(rand.Reader, uuid)
	// if n != len(uuid) || err != nil {
	// 	return "", err
	// }
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
