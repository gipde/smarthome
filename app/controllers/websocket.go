package controllers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
	"schneidernet/smarthome/app/dao"
	"schneidernet/smarthome/app/models/alexa"
	"schneidernet/smarthome/app/models/devcom"
	"strconv"
	"strings"
	"time"
)

/*
Das Websocket empfängt Nachrichten von den Clients und sendet wieder
welche zurück.
Das Rendering ist grundsätzlich Aufgabe des Clients selbst

*/

// check if user has permission for websocket
// only if no session is available
func (c Main) checkWebsocketBasicAuth() revel.Result {
	auth := c.Request.Header.Get("Authorization")
	if auth != "" {

		// Split up the string to get just the data, then get the credentials
		username, password, err := getCredentials(strings.Split(auth, " ")[1])
		if err != nil {
			return c.RenderError(err)
		}

		dbUsr := dao.GetUser(username)
		if dbUsr != nil {
			err := bcrypt.CompareHashAndPassword(dbUsr.DevicePassword, []byte(password))
			if err == nil {
				c.setSession(strconv.Itoa(int(dbUsr.ID)), dbUsr.UserID)
				return nil
			}
		}
		revel.AppLog.Error("WS-Auth - User not found", "user", username)
	}

	revel.AppLog.Error("WS-Auth Error")
	return c.RenderError(errors.New("401: Not authorized"))
}

// DeviceFeed is the Main-Entry for Websocket
func (c Main) DeviceFeed(ws revel.ServerWebSocket) revel.Result {

	usertopic, consumer := register(c.getCurrentUserID())
	c.Log.Debugf("Starting Websocket: %+v", usertopic)

	// start consumer-handler
	consumerHandler(ws, consumer, c.getCurrentUserID())

	// connection state
	primaryConnector := make(map[uint]bool)

	//external Receiver from Websocket
	for {
		var msg string
		err := ws.MessageReceiveJSON(&msg)
		c.Log.Debugf("Received Message %+v -> %s", ws, msg)
		if err != nil {
			c.Log.Debugf("we got a error on Receiving from Websocket %v", err)
			break
		}

		var incoming devcom.DevProto
		err = json.Unmarshal([]byte(msg), &incoming)
		if err != nil {
			c.Log.Errorf("Error in conversion %v", err)
		}

		switch incoming.Action {

		case devcom.ListeDevices:
			devices := dao.GetAllDevices(c.getCurrentUserID())
			devicesView := make([]devcom.Device, len(*devices))
			for i, d := range *devices {
				devicesView[i].ID = d.FQID()
				devicesView[i].Name = d.Name
				devicesView[i].Description = d.Description
				devicesView[i].Connected = d.Connected
				devicesView[i].DeviceType = d.DeviceType
			}
			ws.MessageSendJSON(devcom.DevProto{
				Action: devcom.DeviceList,
				PayLoad: struct{ Devices []devcom.Device }{
					Devices: devicesView,
				},
			})

		case devcom.Ping:
			ws.MessageSendJSON(devcom.DevProto{
				Action: devcom.Pong,
				PayLoad: struct{ Time time.Time }{
					Time: time.Now(),
				},
			})

		case devcom.Pong:
			// TODO: if no Pong will be received ->
			// terminate consumer-hdl, websocket, and so on

		case devcom.RequestState:
			dbdev := dao.FindDeviceByID(c.getCurrentUserID(), incoming.Device.ID)
			c.Log.Debug("Send back State")
			ws.MessageSendJSON(convertToDevcom(dbdev, devcom.StateResponse))

		case devcom.SetState:
			setState(&ws, incoming, c.getCurrentUserID(), func(device *dao.Device) bool {
				device.State = fmt.Sprintf("%v", incoming.PayLoad)
				return true
			})

		case devcom.FlipState:
			setState(&ws, incoming, c.getCurrentUserID(), func(device *dao.Device) bool {
				// only for conected devices
				switch device.State {
				case alexa.ON:
					device.State = alexa.OFF
				case alexa.OFF:
					device.State = alexa.ON
				}
				return true
			})

		case devcom.Connect:
			setState(&ws, incoming, c.getCurrentUserID(), func(device *dao.Device) bool {
				device.Connected = true
				if _, ok := primaryConnector[device.ID]; ok {
					c.Log.Error("Device already connected - cannot connect")
					return false
				}
				primaryConnector[device.ID] = true
				return true
			})

		case devcom.Disconnect:
			setState(&ws, incoming, c.getCurrentUserID(), func(device *dao.Device) bool {
				device.Connected = false
				delete(primaryConnector, device.ID)
				dao.SaveDevice(device)
				// send to all Consumers
				notifyAlLConsumer(c.getCurrentUserID(), convertToDevcom(device, devcom.StateUpdate))
				return false
			})
		}

	}

	// loop over all entries in primaryConnecotr (per device)
	// if this was the connetion who has directly connected to the device -> disconnect and notify
	for k, v := range primaryConnector {
		if v {
			c.Log.Debug("Primary Connector closed, inform other")
			dbdev := dao.GetDevice(c.getCurrentUserID(), k)
			dbdev.Connected = false
			notifyAlLConsumer(c.getCurrentUserID(), convertToDevcom(dbdev, devcom.StateUpdate))
			dao.SaveDevice(dbdev)
		}
	}

	unregister(c.getCurrentUserID(), consumer)
	c.Log.Debugf("Quit Websocket: %+v", ws)
	return nil
}

func setState(ws *revel.ServerWebSocket, incoming devcom.DevProto, useroid uint, payloadhdl func(device *dao.Device) bool) {
	dbdev := dao.FindDeviceByID(useroid, incoming.Device.ID)
	// save only if device is connected and hdl returns true
	if payloadhdl(dbdev) && dbdev.Connected {
		dao.SaveDevice(dbdev)
		// send to all Consumers
		notifyAlLConsumer(useroid, convertToDevcom(dbdev, devcom.StateUpdate))
	}
}

func convertToDevcom(dbdev *dao.Device, action string) *devcom.DevProto {
	return &devcom.DevProto{
		Action: action,
		Device: devcom.Device{
			Connected:   dbdev.Connected,
			Description: dbdev.Description,
			DeviceType:  dbdev.DeviceType,
			Name:        dbdev.Name,
			ID:          dbdev.FQID(),
		},
		PayLoad: dbdev.State,
	}
}

func getCredentials(data string) (username, password string, err error) {
	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", "", err
	}
	strData := strings.Split(string(decodedData), ":")
	username = strData[0]
	password = strData[1]
	return
}
