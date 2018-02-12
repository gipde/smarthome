package controllers

import (
	"encoding/base64"
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

	usertopic, consumer := register(c.getCurrentUserID(), c.ClientIP)
	c.Log.Debugf("Starting Websocket: %+v", usertopic)

	// start consumer-handler
	consumerHandler(ws, consumer, c.getCurrentUserID())

	//external Receiver from Websocket
	for {
		var incoming devcom.DevProto
		err := ws.MessageReceiveJSON(&incoming)
		c.Log.Debugf("Received Message %+v -> %s", ws, incoming)
		if err != nil {
			c.Log.Debugf("we got a error on Receiving from Websocket %v", err)
			break
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
				PayLoad: struct {
					Time time.Time
					ID   string
				}{
					Time: time.Now(),
					ID:   consumer.ID,
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
			modify(incoming.Device.ID, c.getCurrentUserID(), func(device *dao.Device) bool {
				device.State = fmt.Sprintf("%v", incoming.PayLoad)
				checkAndSetAutoOff(device)
				dao.PersistLog(device.ID, "Device state: "+device.State)
				return true
			})

		case devcom.FlipState:
			modify(incoming.Device.ID, c.getCurrentUserID(), func(device *dao.Device) bool {
				// only for conected devices
				switch device.State {
				case alexa.ON:
					device.State = alexa.OFF
				case alexa.OFF:
					device.State = alexa.ON
				}
				dao.PersistLog(device.ID, "Device state: "+device.State)
				checkAndSetAutoOff(device)
				return true
			})

		case devcom.Connect:
			modify(incoming.Device.ID, c.getCurrentUserID(), func(device *dao.Device) bool {
				// we accept any (re)connect
				device.Connected = consumer.ID
				newState := fmt.Sprintf("%v", incoming.PayLoad)
				if !(newState == alexa.ON || newState == alexa.OFF) {
					// no state defined
					return false
				}
				device.State = newState
				dao.PersistLog(device.ID, "Device connected: "+consumer.ID)
				return true
			})

		case devcom.Disconnect:
			modify(incoming.Device.ID, c.getCurrentUserID(), func(device *dao.Device) bool {
				dao.PersistLog(device.ID, "Device disconnected normally: "+device.Connected)
				device.Connected = ""
				dao.SaveDevice(device)
				// send to all Consumers
				notifyAlLConsumer(c.getCurrentUserID(), convertToDevcom(device, devcom.StateUpdate))
				return false
			})
		}

	}

	// loop over all devices
	// if this was the connetion who has directly connected to the device -> disconnect and notify
	for _, dbdev := range *dao.GetAllDevices(c.getCurrentUserID()) {
		if dbdev.Connected == consumer.ID {
			dao.PersistLog(dbdev.ID, "Device disconnected abnormally: "+dbdev.Connected)
			dbdev.Connected = ""
			c.Log.Debug("Connector disconnected from ", "device", dbdev.ID, "consumer", consumer.ID)
			notifyAlLConsumer(c.getCurrentUserID(), convertToDevcom(&dbdev, devcom.StateUpdate))
			dao.SaveDevice(&dbdev)
		}
	}

	unregister(c.getCurrentUserID(), consumer)
	c.Log.Debugf("Quit Websocket: %+v", ws)
	return nil
}

func SetStateWithCheckAutoOff(device *dao.Device, state string) {
	SetState(device, state)
	checkAndSetAutoOff(device)
}

func SetState(dbdev *dao.Device, state string) {
	dbdev.State = state
	saveAndNotify(dbdev)
	dao.PersistLog(dbdev.ID, "Device State "+state)
}

func checkAndSetAutoOff(device *dao.Device) {
	if device.AutoCountDown > 0 && device.State == alexa.ON {
		dao.PersistLog(device.ID, "set Countdown (minutes): "+strconv.Itoa(device.AutoCountDown))
		sched := dao.CreateCountDown(time.Now().Add(time.Duration(device.AutoCountDown)*time.Minute), alexa.OFF, device.ID)
		dao.SaveSchedule(sched)
	}
}

func saveAndNotify(dbdev *dao.Device) {
	if dbdev.Connected != "" {
		dao.SaveDevice(dbdev)
		// send to all Consumers
		notifyAlLConsumer(dbdev.UserID, convertToDevcom(dbdev, devcom.StateUpdate))
	}
}

func modify(devName string, useroid uint, callback func(device *dao.Device) bool) {
	dbdev := dao.FindDeviceByID(useroid, devName)
	// save only if device is connected and hdl returns true
	if callback(dbdev) {
		saveAndNotify(dbdev)
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
