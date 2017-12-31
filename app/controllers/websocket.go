package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/revel/revel"
	"schneidernet/smarthome/app/dao"
	"schneidernet/smarthome/app/models/alexa"
	"schneidernet/smarthome/app/models/devcom"
	"strings"
	"time"
)

/*
Das Websocket empfängt Nachrichten von den Clients und sendet wieder
welche zurück.
Das Rendering ist grundsätzlich Aufgabe des Clients selbst

*/
func (c Main) DeviceFeed(ws revel.ServerWebSocket) revel.Result {

	usertopic, consumer := register(c.getCurrentUserID())

	c.Log.Debugf("%+v", usertopic)
	// TODO: set connected-Flag on Websocket Connection from Device not from browser

	consumerHandler(ws, consumer)

	// connection state
	connected := make(map[uint]bool)

	//external Receiver from Websocket
	for {
		var msg string
		err := ws.MessageReceiveJSON(&msg)
		c.Log.Debugf("%+v -> %s", ws, msg)
		if err != nil {
			c.Log.Errorf("we got a error %v", err)
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

		case devcom.RequestState:
			dbdev := dao.FindDeviceByID(c.getCurrentUserID(), incoming.Device.ID)
			ws.MessageSendJSON(convertToDevcom(dbdev, devcom.StateResponse))

		case devcom.SetState:
			setState(&ws, incoming, c.getCurrentUserID(), func(device *dao.Device) bool {
				device.State = fmt.Sprintf("%v", incoming.PayLoad)
				return true
			})

		case devcom.FlipState:
			setState(&ws, incoming, c.getCurrentUserID(), func(device *dao.Device) bool {
				// only for conected devices
				if device.Connected {
					switch device.State {
					case alexa.ON:
						device.State = alexa.OFF
					case alexa.OFF:
						device.State = alexa.ON
					}
					return true
				}
				return false
			})

		case devcom.Connect:
			setState(&ws, incoming, c.getCurrentUserID(), func(device *dao.Device) bool {
				device.Connected = true
				if _, ok := connected[device.ID]; ok {
					c.Log.Error("Device already connected - cannot connect")
					return false
				}
				connected[device.ID] = true
				return true
			})

		case devcom.Disconnect:
			setState(&ws, incoming, c.getCurrentUserID(), func(device *dao.Device) bool {
				device.Connected = false
				delete(connected, device.ID)
				return true
			})
		}

		if err != nil {
			break
		}
	}

	// if this was the connetion how has directly connected to the device -> disconnect and notify
	for k, v := range connected {
		if v {
			dbdev := dao.GetDevice(c.getCurrentUserID(), k)
			dbdev.Connected = false
			topics[c.getCurrentUserID()].Input <- *convertToDevcom(dbdev, devcom.StateUpdate)
			dao.SaveDevice(dbdev)
		}
	}

	unregister(c.getCurrentUserID(), consumer)
	return nil
}

func setState(ws *revel.ServerWebSocket, incoming devcom.DevProto, useroid uint, payloadhdl func(device *dao.Device) bool) {
	dbdev := dao.FindDeviceByID(useroid, incoming.Device.ID)
	if payloadhdl(dbdev) {
		dao.SaveDevice(dbdev)
		// send to all Consumers
		topics[useroid].Input <- *convertToDevcom(dbdev, devcom.StateUpdate)
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
