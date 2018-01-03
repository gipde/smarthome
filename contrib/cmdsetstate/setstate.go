package main

import (
	"encoding/base64"
	"flag"
	"golang.org/x/net/websocket"
	"log"
	"schneidernet/smarthome/app/models/devcom"
)

/*
go run setstate.go -url ws://localhost:8180/Main/DeviceFeed -user admin -pass admin -device device-1 -state OFF

*/

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func main() {

	url := flag.String("url", "wss://2e1512f0-d590-4eed-bb41-9ad3abd03edf.pub.cloud.scaleway.com/sh/Main/DeviceFeed", "websocket url")
	user := flag.String("user", "", "username for accessing device")
	pass := flag.String("pass", "", "password for accessing device")
	dev := flag.String("device", "", "device")
	state := flag.String("state", "", "state")
	flag.Parse()

	log.Println("Connecting to" + *url)

	config, _ := websocket.NewConfig(*url, "http://localhost/")
	config.Header.Add("Authorization", "Basic "+basicAuth(*user, *pass))
	ws, err := websocket.DialConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	sendToWebsocket(ws, &devcom.DevProto{
		Action: devcom.Connect,
		Device: devcom.Device{
			ID: *dev,
		},
	})

	sendToWebsocket(ws, &devcom.DevProto{
		Action: devcom.SetState,
		Device: devcom.Device{
			ID: *dev,
		},
		PayLoad: state,
	})

}

func sendToWebsocket(ws *websocket.Conn, command *devcom.DevProto) {
	log.Println("Sending...")
	err := websocket.JSON.Send(ws, command)
	if err != nil {
		log.Fatal(err)
	}

	var incoming interface{}
	err = websocket.JSON.Receive(ws, &incoming)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("response: %+v\n", incoming)

}
