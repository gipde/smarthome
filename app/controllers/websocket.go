package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/revel/revel"
	"schneidernet/smarthome/app/dao"
	"schneidernet/smarthome/app/models/alexa"
	"schneidernet/smarthome/app/models/websocket"
	"strings"
)

// Global Topics-Map per useroid
var topics = make(map[uint]*websocket.StateTopic)

// register new User for his topic
func register(uoid uint) (chan string, chan string) {
	if _, ok := topics[uoid]; !ok {
		// we create a new StateTopic
		topic := websocket.StateTopic{
			Input:    make(chan string),
			Consumer: [](chan string){},
		}
		topics[uoid] = &topic

		// and start a per user TopicHandler
		topicHandler(&topic)
	}

	usertopic := topics[uoid]
	// we add a consumer
	consumer := make(chan string)
	usertopic.Consumer = append(usertopic.Consumer, consumer)

	return usertopic.Input, consumer
}

// unregister user from his topic
func unregister(uoid uint, consumer chan string) {
	usertopic := topics[uoid]

	for i, c := range usertopic.Consumer {
		if c == consumer {
			// send the consumer the Quit Command and close it
			c <- "QUIT"
			close(c)

			// remove the Consumer
			usertopic.Consumer = append(usertopic.Consumer[:i], usertopic.Consumer[i+1:]...)
		}
	}

	// if this was the last consumer
	if len(usertopic.Consumer) == 0 {
		// we can send quit to the usertopic and close the usertopic
		usertopic.Input <- "QUIT"
		close(usertopic.Input)
		// delete the complete usertopic
		delete(topics, uoid)
	}
}

// the topicHandler
func topicHandler(stateTopic *websocket.StateTopic) {
	// start goroutine and loop forever
	go func() {
		for {
			msg := <-stateTopic.Input
			if msg == "QUIT" {
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
func consumerHandler(ws revel.ServerWebSocket, consumer chan string) {
	//internal Receiver from StateTopic loop forever
	go func() {
		for {
			msg := <-consumer
			// after here, it is possible that WebSocketController is disabled
			if msg == "QUIT" {
				break
			}

			// send to Websocket
			var devState = websocket.DeviceCommand{}
			json.Unmarshal([]byte(msg), &devState)
			err := ws.MessageSendJSON(&devState)

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

	// TODO: set connected-Flag on Websocket Connection from Device not from browser

	consumerHandler(ws, consumer)

	//external Receiver from Websocket
	for {
		var msg string
		err := ws.MessageReceiveJSON(&msg)
		if err != nil {
			c.Log.Errorf("we got a error %v", err)
			break
		}

		var incoming websocket.DeviceCommand
		err = json.Unmarshal([]byte(msg), &incoming)
		if err != nil {
			c.Log.Errorf("Error in conversion %v", err)
		}

		dev := dao.FindDeviceByID(c.getCurrentUserID(), incoming.Device)
		incoming.DeviceType = dev.DeviceType

		switch incoming.Command {
		case "CLICK":
			protocolClick(usertopic, &incoming, dev)
		case "SETSTATE":
			protocolSetState(usertopic, &incoming, dev)
		case "GETSTATE":
			err = protocolGetState(&incoming, dev, ws)
		}

		if err != nil {
			break
		}
	}

	unregister(c.getCurrentUserID(), consumer)
	return nil
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

// Notify a state Change
func notifyStateUpdate(user uint, device *dao.Device) {
	devname := fmt.Sprintf("device-%d", device.ID)
	dc := websocket.DeviceCommand{
		Device:     devname,
		Connected:  device.Connected,
		Command:    "STATEUPDATE",
		State:      device.State,
		DeviceType: device.DeviceType,
	}
	data, _ := json.Marshal(dc)
	topics[user].Input <- string(data)
}

func protocolClick(usertopic chan string, cmd *websocket.DeviceCommand, dev *dao.Device) {
	cmd.Connected = dev.Connected
	if cmd.Connected {
		switch alexa.DeviceType(dev.DeviceType) {
		case alexa.DeviceLight,
			alexa.DeviceSocket,
			alexa.DeviceSwitch:

			if dev.State == "OFF" {
				dev.State = "ON"
			} else {
				dev.State = "OFF"
			}

			// if we have to send back
			dao.SaveDevice(dev)
			cmd.State = dev.State
			cmd.Command = "STATEUPDATE"

			j, _ := json.Marshal(&cmd)

			// send msg internally
			usertopic <- string(j)

		}
	}
}

func protocolSetState(usertopic chan string, cmd *websocket.DeviceCommand, dev *dao.Device) {
	// entweder ist das geraet schon connected, oder es wird gerade verbunden
	if dev.Connected || cmd.Connected {

		// save state in db
		dev.State = cmd.State
		dev.Connected = cmd.Connected
		dao.SaveDevice(dev)

		// send msg internally
		cmd.Command = "STATEUPDATE"
		cmd.Connected = dev.Connected

		data, _ := json.Marshal(&cmd)
		usertopic <- string(data)
	}
}

func protocolGetState(cmd *websocket.DeviceCommand, dev *dao.Device, ws revel.ServerWebSocket) error {
	cmd.Command = "STATERESPONSE"
	cmd.State = dev.State
	cmd.Connected = dev.Connected
	err := ws.MessageSendJSON(&cmd)
	if err != nil {
		return err
	}
	return nil
}
