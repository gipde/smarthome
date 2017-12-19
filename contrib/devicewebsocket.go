package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"os"
	"strings"
)

type DeviceCommand struct {
	Device    string
	Connected bool
	Command   string
	State     string
}

func doSwitch(ws *websocket.Conn, t int) error {
	req := DeviceCommand{
		Device: "device-1",
	}

	switch t {
	case 0:
		req.Command = "GETSTATE"
	case 1:
		req.Command = "SETSTATE"
		req.State = "ON"
	case 2:
		req.Command = "SETSTATE"
		req.State = "OFF"
	}
	fmt.Printf("Send: %+v\n", req)
	return websocket.JSON.Send(ws, &req)
}
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func main() {

	// var host="wss://2e1512f0-d590-4eed-bb41-9ad3abd03edf.pub.cloud.scaleway.com/sh/Main/DeviceFeed"
	var host = "ws://localhost:8180/Main/DeviceFeed"
	config, _ := websocket.NewConfig(host, "http://localhost/")
	config.Header.Add("Authorization", "Basic "+basicAuth("admin", "admin"))
	ws, err := websocket.DialConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			var msg = DeviceCommand{}
			err := websocket.JSON.Receive(ws, &msg)
			if err != nil {
				fmt.Printf("Error: %+v", err)
				break
			}
			fmt.Printf("Received: %s.\n", msg)
		}
		fmt.Printf("leaving listener...\n")
	}()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')

		exit := false
		var err error

		switch strings.ToLower(strings.Replace(text, "\n", "", -1)) {
		case "quit":
			fmt.Println("we quit")
			exit = true
		case "state":
			err = doSwitch(ws, 0)
		case "on":
			err = doSwitch(ws, 1)
		case "off":
			err = doSwitch(ws, 2)
		}

		if exit || err != nil {
			break
		}

	}

	ws.Close()

}
