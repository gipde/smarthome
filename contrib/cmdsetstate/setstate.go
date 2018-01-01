package main

import (
	"encoding/base64"
	"flag"
	"golang.org/x/net/websocket"
	"log"
	"schneidernet/smarthome/app/models/devcom"
	// "time"
)

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

	println("Connecting to" + *url)
	config, _ := websocket.NewConfig(*url, "http://localhost/")
	config.Header.Add("Authorization", "Basic "+basicAuth(*user, *pass))
	ws, err := websocket.DialConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()
	log.Println("Sending Connect...")
	err = websocket.JSON.Send(ws, &devcom.DevProto{
		Action: devcom.Connect,
		Device: devcom.Device{
			ID: *dev,
		},
	})
	var incoming interface{}
	err = websocket.JSON.Receive(ws, &incoming)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("incoming: %+v\n", incoming)
	// time.Sleep(1 * time.Second)
	log.Println("Sending State...")
	err = websocket.JSON.Send(ws, &devcom.DevProto{
		Action: devcom.SetState,
		Device: devcom.Device{
			ID: *dev,
		},
		PayLoad: state,
	})
	if err != nil {
		log.Fatal(err)
	}

	// err = websocket.JSON.Receive(ws, &incoming)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	log.Printf("incoming: %+v\n", incoming)

	// time.Sleep(1 * time.Second)

}
