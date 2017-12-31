package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"schneidernet/smarthome/app/models/alexa"
	"time"
	// "fmt"
	"github.com/revel/revel"
	"schneidernet/smarthome/app/dao"
	// "schneidernet/smarthome/app/models/alexa"
	"schneidernet/smarthome/app/models/devcom"
	"strings"
)

// Internal Message Consumer
type StateTopic struct {
	Input    chan devcom.DevProto
	Consumer [](chan devcom.DevProto)
}

// Global Topics-Map per useroid
var topics = make(map[uint]*StateTopic)

// register new User for his topic
func register(uoid uint) (chan devcom.DevProto, chan devcom.DevProto) {
	if _, ok := topics[uoid]; !ok {
		// we create a new StateTopic
		topic := StateTopic{
			Input:    make(chan devcom.DevProto),
			Consumer: [](chan devcom.DevProto){},
		}
		topics[uoid] = &topic

		// and start a per user TopicHandler
		topicHandler(&topic)
	}

	usertopic := topics[uoid]
	// we add a consumer
	consumer := make(chan devcom.DevProto)
	usertopic.Consumer = append(usertopic.Consumer, consumer)

	return usertopic.Input, consumer
}

// unregister user from his topic
func unregister(uoid uint, consumer chan devcom.DevProto) {
	usertopic := topics[uoid]

	for i, c := range usertopic.Consumer {
		if c == consumer {
			// send the consumer the Quit Command and close it
			c <- devcom.DevProto{Action: devcom.Quit}
			close(c)

			// remove the Consumer
			usertopic.Consumer = append(usertopic.Consumer[:i], usertopic.Consumer[i+1:]...)
		}
	}

	// if this was the last consumer
	if len(usertopic.Consumer) == 0 {
		// we can send quit to the usertopic and close the usertopic
		usertopic.Input <- devcom.DevProto{Action: devcom.Quit}
		close(usertopic.Input)
		// delete the complete usertopic
		delete(topics, uoid)
	}
}

// the topicHandler
func topicHandler(stateTopic *StateTopic) {
	// start goroutine and loop forever
	go func() {
		for {
			msg := <-stateTopic.Input
			if msg.Action == devcom.Quit {
				// exit loop
				break
			}
			// send to every consumer
			for _, consumer := range stateTopic.Consumer {
				consumer <- msg
			}
		}
	}()
}

// the consumerHandler for every Consumer
func consumerHandler(ws revel.ServerWebSocket, consumer chan devcom.DevProto) {
	//internal Receiver from StateTopic loop forever
	go func() {
		for {
			msg := <-consumer
			// after here, it is possible that WebSocketController is disabled
			if msg.Action == devcom.Quit {
				break
			}

			// send to Websocket
			// var devState = devcom.DevProto{}
			// json.Unmarshal([]byte(msg), &devState)
			err := ws.MessageSendJSON(&msg)

			if err != nil {
				break
			}
		}
	}()
}

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

		// dev := dao.FindDeviceByID(c.getCurrentUserID(), incoming.Device.Name)
		// incoming.DeviceType = dev.DeviceType

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
			setState(&ws, incoming, c.getCurrentUserID(), func(device *dao.Device) {
				device.State = fmt.Sprintf("%v", incoming.PayLoad)
			})

		case devcom.FlipState:
			setState(&ws, incoming, c.getCurrentUserID(), func(device *dao.Device) {
				switch device.State {
				case alexa.ON:
					device.State = alexa.OFF
				case alexa.OFF:
					device.State = alexa.ON
				}
			})

		case devcom.Connect:
			setState(&ws, incoming, c.getCurrentUserID(), func(device *dao.Device) {
				device.Connected = true
			})

		case devcom.Disconnect:
			setState(&ws, incoming, c.getCurrentUserID(), func(device *dao.Device) {
				device.Connected = false
			})
		}

		if err != nil {
			break
		}
	}

	unregister(c.getCurrentUserID(), consumer)
	return nil
}

func setState(ws *revel.ServerWebSocket, incoming devcom.DevProto, useroid uint, payloadhdl func(device *dao.Device)) {
	dbdev := dao.FindDeviceByID(useroid, incoming.Device.ID)
	payloadhdl(dbdev)

	dao.SaveDevice(dbdev)
	// send to all Consumers
	topics[useroid].Input <- *convertToDevcom(dbdev, devcom.StateUpdate)

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
