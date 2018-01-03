package main

import (
	"encoding/base64"
	"flag"
	"log"
	"os"
	"os/signal"
	"schneidernet/smarthome/app/models/devcom"
	"time"

	"github.com/stianeikeland/go-rpio"
	"golang.org/x/net/websocket"
)

/*
Compile for Linux
env GOOS=linux GOARCH=arm GOARM=6 go build

Start examples
./rpiswitch -url ws://jupiter:9000/Main/DeviceFeed -device device-1 -user admin -pass admin -port 4
./rpiswitch -url ws://jupiter:8180/Main/DeviceFeed -device device-5 -user gipde90@gmail.com -pass hallo -port 4
./rpiswitch -device device-5 -user gipde90@gmail.com -pass hallo -port 4


Test with sysfs

cd /sys/class/gpio
echo 4 > export

echo in > gpio4/direction
cat gpio4/value

echo out > gpio4/direction

echo 1 > gpio4/value
echo 0 > gpio4/value

echo 4 > unexport

*/

var ctrlC = false

const (
	On  = "ON"
	Off = "OFF"
)

var (
	url, user, pass, dev *string
	port                 *int
	test, checkstate     *bool

	currentState string
	wsErr        error
)

func main() {

	pin := checkArgs()

	// Marker -> to see that program is working
	go markHandler()
	// Ctrl-C Handler
	ctrlCHandler()

	// Mainloop with Reconnect on Error
	for {
		wsErr = nil
		startWebsocket(pin)

		time.Sleep(time.Second * 1)
		log.Println("Reconnect ...")
	}
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func startWebsocket(pin rpio.Pin) {
	log.Println("Connecting to " + *url)

	config, _ := websocket.NewConfig(*url, "http://localhost/")
	config.Header.Add("Authorization", "Basic "+basicAuth(*user, *pass))
	ws, err := websocket.DialConfig(config)
	if err != nil {
		wsErr = err
		return
	}

	// websocket receive handler
	go websocketHandler(ws, pin)

	// connect
	sendToWebsocket(ws, &devcom.DevProto{
		Action: devcom.Connect,
		Device: devcom.Device{
			ID: *dev,
		},
	})

	// .. and transmit current state
	currentState = getState(pin)
	transmitState(currentState, ws, dev)

	// Loop Forever
	for {
		if *checkstate {
			newstate := getState(pin)
			log.Printf("State from switch: %s > from server %s\n", newstate, currentState)
			if newstate != currentState {
				log.Println("NEW STATE FROM SWITCH: " + newstate)
				transmitState(newstate, ws, dev)
				currentState = newstate
			}
		}
		//error on websocket
		if wsErr != nil {
			return
		}
		if ctrlC {
			cleanUpAndExit(ws, dev)
		}

		time.Sleep(1 * time.Second)
	}
}

func websocketHandler(ws *websocket.Conn, pin rpio.Pin) {
	for {
		var incoming devcom.DevProto
		err := websocket.JSON.Receive(ws, &incoming)
		if err != nil {
			wsErr = err
			log.Printf("Error: %+v", err)
			return
		}
		log.Printf("we received: %+v\n", incoming)
		switch incoming.Action {
		// stateupdate received
		case devcom.StateUpdate:

			switch incoming.PayLoad {
			case On:
				setState(pin, On)
			case Off:
				setState(pin, Off)
			}

		}
	}
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

func sendToWebsocket(ws *websocket.Conn, command *devcom.DevProto) {
	log.Printf("Sending %+v ...\n", command)
	err := websocket.JSON.Send(ws, command)
	if err != nil {
		wsErr = err
		log.Printf("Error: %+v", err)
	}
}

func openPin(pin uint) rpio.Pin {
	log.Printf("Opening RPI Port %d\n", pin)
	err := rpio.Open()
	if err != nil {
		log.Fatal("Error opening", "port", pin)
	}
	result := rpio.Pin(pin)
	result.Output()

	return result
}

func getState(pin rpio.Pin) string {
	x := pin.Read()
	var state string
	if x == 1 {
		state = Off
	} else {
		state = On
	}
	return state
}

func setState(pin rpio.Pin, state string) {
	log.Println("Switch " + state)
	if state == On {
		pin.Low()
	} else {
		pin.High()
	}
	currentState = state
}

func checkArgs() rpio.Pin {
	url = flag.String("url", "wss://2e1512f0-d590-4eed-bb41-9ad3abd03edf.pub.cloud.scaleway.com/sh/Main/DeviceFeed", "websocket url")
	user = flag.String("user", "", "username for accessing device")
	pass = flag.String("pass", "", "password for accessing device")
	dev = flag.String("device", "", "device name (e.g. device-1)")
	port = flag.Int("port", 0, "raspberr-pi gpio port")
	test = flag.Bool("test", false, "test gpio port")
	checkstate = flag.Bool("checkstate", false, "periodically check state of device and notify change of state")
	flag.Parse()

	exitErr(*port == 0, "Port not set")
	pin := openPin(uint(*port))
	if *test {

		fn := func(s string) {
			log.Println(s)
			setState(pin, s)
			log.Println("Current State: " + getState(pin) + "\n")
			time.Sleep(5 * time.Second)
		}
		for _, s := range []string{On, Off, On, Off} {
			fn(s)
		}
		os.Exit(0)
	}

	exitErr(*url == "", "URL not set")
	exitErr(*user == "", "User not set")
	exitErr(*pass == "", "Password not set")
	exitErr(*dev == "", "Device not set")

	return pin
}

func exitErr(cond bool, err string) {
	if cond {
		flag.PrintDefaults()
		log.Fatal(err)
	}
}

func cleanUpAndExit(ws *websocket.Conn, dev *string) {
	// send disconnect
	sendToWebsocket(ws, &devcom.DevProto{
		Action: devcom.Disconnect,
		Device: devcom.Device{
			ID: *dev,
		},
	})

	//chance to get response
	time.Sleep(1 * time.Second)
	ws.Close()

	//finally exit
	os.Exit(0)
}

func markHandler() {
	for {
		log.Println("--MARK--")
		time.Sleep(1 * time.Hour)
	}
}

func ctrlCHandler() {
	c := make(chan os.Signal, 1)
	signal.Reset(os.Interrupt)
	signal.Notify(c, os.Interrupt)
	go func() {
		//receive msg
		<-c

		log.Println("we received CTRL+C")
		ctrlC = true
		// exit hard if websocket hangs
		go func() {
			time.Sleep(5 * time.Second)
			os.Exit(0)
		}()

	}()
}
