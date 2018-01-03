package main

import (
	"encoding/base64"
	"flag"
	"github.com/stianeikeland/go-rpio"
	"golang.org/x/net/websocket"
	"log"
	"os"
	"os/signal"
	"schneidernet/smarthome/app/models/devcom"
	"strconv"
	"time"
)

/*
env GOOS=linux GOARCH=arm GOARM=6 go build

./rpiswitch -url ws://jupiter:9000/Main/DeviceFeed -user admin -pass admin -device device-1

*/

var done = false

func main() {
	url := flag.String("url", "wss://2e1512f0-d590-4eed-bb41-9ad3abd03edf.pub.cloud.scaleway.com/sh/Main/DeviceFeed", "websocket url")
	user := flag.String("user", "", "username for accessing device")
	pass := flag.String("pass", "", "password for accessing device")
	dev := flag.String("device", "", "device")
	port := flag.String("port", "", "port")
	flag.Parse()
	checkArgs(url, user, pass, dev, port)

	log.Println("Connecting to " + *url)

	config, _ := websocket.NewConfig(*url, "http://localhost/")
	config.Header.Add("Authorization", "Basic "+basicAuth(*user, *pass))
	ws, err := websocket.DialConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// RPI Pin
	portI, _ := strconv.Atoi(*port)
	pin := openPin(uint(portI))

	// Cleanup
	initCleanupHandler(ws, dev)

	// websocket receive handler
	go websocketHandler(ws, pin)

	// connect
	sendToWebsocket(ws, &devcom.DevProto{
		Action: devcom.Connect,
		Device: devcom.Device{
			ID: *dev,
		},
	})

	// Marker -> to see that program is working
	go func() {
		for {
			log.Println("--MARK--")
			time.Sleep(1 * time.Hour)
		}
	}()

	// .. and transmit current state
	currentState := getState(pin)
	transmitState(currentState, ws, dev)

	// loop forever and read every second current state
	for {
		newstate := getState(pin)
		log.Print("State from switch: ")
		log.Println(newstate)
		if newstate != currentState {
			log.Println("NEW STATE FROM SWITCH: " + newstate)
			transmitState(newstate, ws, dev)
			currentState = newstate
		}
		time.Sleep(1 * time.Second)
	}
}

func checkArgs(args ...*string) {
	for _, arg := range args {
		if *arg == "" {
			flag.PrintDefaults()
			log.Fatal("Argument not set")
		}
	}
}

func initCleanupHandler(ws *websocket.Conn, dev *string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Println("we received CTRL+C")
		sendToWebsocket(ws, &devcom.DevProto{
			Action: devcom.Disconnect,
			Device: devcom.Device{
				ID: *dev,
			},
		})
		done = true
	}()
}

func websocketHandler(ws *websocket.Conn, pin rpio.Pin) {
	for {
		var incoming devcom.DevProto
		err := websocket.JSON.Receive(ws, &incoming)
		if err != nil {
			log.Printf("Error: %+v", err)
			break
		}
		log.Printf("we received: %+v\n", incoming)
		switch incoming.Action {
		case devcom.StateUpdate:
			pin.Output()
			switch incoming.PayLoad {
			case "ON":
				log.Println("SET STATE HIGH -> OFF")
				pin.High()
			case "OFF":
				log.Println("SET STATE LOW -> ON")
				pin.Low()
			}
		}
		if done {
			ws.Close()
			os.Exit(0)
		}
	}
	log.Println("Ending WS Incoming Handler")

}

func getState(pin rpio.Pin) string {
	pin.Input()
	var state string
	if pin.Read() == 1 {
		state = "OFF"
	} else {
		state = "ON"
	}
	return state
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func transmitState(state string, ws *websocket.Conn, dev *string) {
	sendToWebsocket(ws, &devcom.DevProto{
		Action: devcom.SetState,
		Device: devcom.Device{
			ID: *dev,
		},
		PayLoad: state,
	})
}

func openPin(pin uint) rpio.Pin {
	log.Printf("Opening RPI Port %d\n", pin)
	err := rpio.Open()
	if err != nil {
		log.Fatal("Error opening", "port", pin)
	}
	return rpio.Pin(pin)
}

func sendToWebsocket(ws *websocket.Conn, command *devcom.DevProto) {
	log.Printf("Sending %+v ...\n", command)
	err := websocket.JSON.Send(ws, command)
	if err != nil {
		log.Fatal(err)
	}
}
