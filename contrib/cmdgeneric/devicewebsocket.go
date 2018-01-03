package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/net/websocket"
	"log"
	"os"
	"schneidernet/smarthome/app/models/devcom"
	. "schneidernet/smarthome/app/models/devcom"
	"strings"
	"time"
)

//go run devicewebsocket.go -user admin -pass admin -url ws://localhost:8180/Main/DeviceFeed

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func main() {

	url := flag.String("url", "wss://2e1512f0-d590-4eed-bb41-9ad3abd03edf.pub.cloud.scaleway.com/sh/Main/DeviceFeed", "websocket url")
	user := flag.String("user", "", "username for accessing device")
	pass := flag.String("pass", "", "password for accessing device")

	flag.Parse()

	config, _ := websocket.NewConfig(*url, "http://localhost/")
	config.Header.Add("Authorization", "Basic "+basicAuth(*user, *pass))
	ws, err := websocket.DialConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			var incoming devcom.DevProto
			err := websocket.JSON.Receive(ws, &incoming)
			if err != nil {
				fmt.Printf("Error: %+v", err)
				break
			}
			log.Printf("Received: %+v.\n", incoming)

			switch incoming.Action {
			case devcom.Ping:
				handle(incoming, pong)
			case devcom.DeviceList:
				handle(incoming, deviceList)
			case devcom.StateResponse:
				handle(incoming, stateResponse)
			}

		}
		fmt.Printf("leaving listener...\n")
	}()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')

		exit := false
		var err error

		cmd := strings.Replace(text, "\n", "", -1)

		switch {
		case cmd == "quit":
			fmt.Println("we quit")
			exit = true
		case cmd == "devices":
			websocket.JSON.Send(ws, &DevProto{
				Action: ListeDevices,
			})
		case cmd == "ping":
			websocket.JSON.Send(ws, &DevProto{
				Action: Ping,
				PayLoad: struct{ Time time.Time }{
					Time: time.Now(),
				},
			})
		case strings.HasPrefix(cmd, "getstate "):
			device := strings.TrimPrefix(cmd, "getstate ")
			websocket.JSON.Send(ws, &DevProto{
				Action: RequestState,
				Device: Device{
					ID: device,
				},
			})
		case strings.HasPrefix(cmd, "setstate "):
			payload := strings.TrimPrefix(cmd, "setstate ")
			splitted := strings.Fields(payload)

			websocket.JSON.Send(ws, &DevProto{
				Action: SetState,
				Device: Device{
					ID: splitted[0],
				},
				PayLoad: splitted[1],
			})
		case strings.HasPrefix(cmd, "connect "):
			device := strings.TrimPrefix(cmd, "connect ")
			websocket.JSON.Send(ws, &DevProto{
				Action: Connect,
				Device: Device{
					ID: device,
				},
			})
		case strings.HasPrefix(cmd, "disconnect "):
			device := strings.TrimPrefix(cmd, "disconnect ")
			websocket.JSON.Send(ws, &DevProto{
				Action: Disconnect,
				Device: Device{
					ID: device,
				},
			})
		}

		if exit || err != nil {
			break
		}

	}

	ws.Close()

}

func handle(r devcom.DevProto, fn func(devcom.DevProto)) {
	fn(r)
}

func pong(r devcom.DevProto) {
	println("we got a ping -> send a pong")
}

func deviceList(r devcom.DevProto) {
	p := struct{ Devices []devcom.Device }{}
	mapstructure.Decode(r.PayLoad, &p)
	for _, d := range p.Devices {
		log.Printf("%s, %s, %s, %d\n", d.ID, d.Name, d.Description, d.DeviceType)
	}
}

func stateResponse(r devcom.DevProto) {
	log.Printf("DEVICE: %+v\n", r.Device)
	log.Printf("STATE: %+v\n", r.PayLoad)
}
