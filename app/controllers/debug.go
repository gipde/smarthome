package controllers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
	"runtime/pprof"
	"schneidernet/smarthome/app"
	"schneidernet/smarthome/app/dao"
	"schneidernet/smarthome/app/models/alexa"
	"schneidernet/smarthome/app/routes"
	"strconv"
	"time"
)

// Debug Controller
type Debug struct {
	*revel.Controller
}

// ListTokens list all Tokens that are issued
func (c Debug) ListTokens() revel.Result {
	html := "<table style='width:100%;border: 1px solid black;'>"
	tokens := dao.GetAllTokens()
	for _, t := range *tokens {
		html += "<tr>"
		html += "<td style='border: 1px solid black;'>" + strconv.Itoa(int(t.ID)) + "</td>"
		html += "<td style='border: 1px solid black;'>" + t.Signature + "</td>"
		html += "<td style='border: 1px solid black;'>" + t.TokenID + "</td>"
		html += "<td style='border: 1px solid black;'>" + string(t.TokenType) + "</td>"
		html += "<td style='border: 1px solid black;'>" + t.Expiry.Format(time.RFC3339) + "</td>"
		html += "<td style='border: 1px solid black;'>" + string(t.PayLoad) + "</td>"
		html += "</tr>"
	}
	return c.RenderHTML(html)
}

// CheckToken verifies the validity of a token
func (c Debug) CheckToken(token string) revel.Result {
	valid, user := CheckToken(token)
	return c.RenderText(fmt.Sprintf("active: %t\nuser: %s\n", valid, user))
}

// GetHash generates a Hash from a Password
func (c Debug) GetHash(password string) revel.Result {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	retval := struct{ Password []byte }{Password: hash}
	return c.RenderJSON(retval)
}

// BasicAuthEncode Encodes a username and password to base64
func (c Debug) BasicAuthEncode(username, password string) revel.Result {
	auth := username + ":" + password
	return c.RenderText(base64.StdEncoding.EncodeToString([]byte(auth)))
}

// BasicAuthDecode reveals the username and password
func (c Debug) BasicAuthDecode(credential string) revel.Result {
	payload, _ := base64.StdEncoding.DecodeString(credential)
	c.Log.Infof("Payload %s", payload)
	return c.RenderText(string(payload))
}

// LogRequest pretty print a request
func (c Debug) LogRequest() revel.Result {
	DoLogRevelRequest(c.Request, "LoggingAPI Endpoint")
	return c.NotFound("but your Request is Logged :)")
}

// DoLogRevelRequest ...
func DoLogRevelRequest(req *revel.Request, prefix string) {
	originalHeader := req.Header.Server.(*revel.GoHeader)
	r := originalHeader.Source.(*revel.GoRequest).Original
	app.DoLogHTTPRequest(r, prefix)
}

// Create Testdevices
func (c Main) CreateDev() revel.Result {

	usr := dao.GetUserWithID(c.getCurrentUserID())
	if usr == nil {
		c.Log.Info("creating users")
		return nil
	}
	usr.Devices = []dao.Device{}

	for i := 0; i < 10; i++ {
		d := dao.Device{
			Name:        "Schalter",
			Description: "Schalter im Flur",
			Producer:    "Werner@SchneiderNET",
			DeviceType:  alexa.DeviceSwitch.ID(),
		}
		usr.Devices = append(usr.Devices, d)
	}
	dao.SaveUser(usr)
	return c.Redirect(app.ContextRoot + routes.Main.Dashboard())
}

func (c Main) Debug() revel.Result {
	return c.Render()
}

func (c Debug) ListGoroutines() revel.Result {
	buf := new(bytes.Buffer)
	pprof.Lookup("goroutine").WriteTo(buf, 1)
	return c.RenderText(buf.String())
}

func (c Debug) ListStateTopic() revel.Result {
	result := fmt.Sprintf("Topics Overall: %d\n", len(topics))
	for k, v := range topics {
		result += fmt.Sprintf("Topic for user %d = %+v\n", k, v.Input)
		result += fmt.Sprintf("Consumer: \n")
		for i, c := range v.Consumer {
			result += fmt.Sprintf("  %d. Consumer: %+v\n", i, c)
		}
		result += "\n"
	}
	return c.RenderText(result)
}

func (c Debug) SetState(device uint, state string, connected string) revel.Result {
	c.Log.Debug("set state", "device", device, "state", state, "connected", connected)
	dbdev := dao.GetDeviceById(device)
	dbdev.State = state
	dbdev.Connected = connected
	saveAndNotify(dbdev)

	return c.Redirect(routes.Main.Debug())
}
