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
env GOOS=linux GOARCH=arm GOARM=6 go build

./rpiswitch -url ws://jupiter:9000/Main/DeviceFeed -user admin -pass admin -device device-1


sysfs

cd /sys/class/gpio
echo 4 > export

echo in > gpio4/direction
cat gpio4/value

echo out > gpio4/direction

echo 1 > gpio4/value
echo 0 > gpio4/value

echo 4 > unexport

*/

var done = false

const (
	On  = "ON"
	Off = "OFF"
)

var (
	url, user, pass, dev *string
	port                 *int
	test, checkstate     *bool

	exitHdl, endWS chan bool
)

func main() {
	url = flag.String("url", "wss://2e1512f0-d590-4eed-bb41-9ad3abd03edf.pub.cloud.scaleway.com/sh/Main/DeviceFeed", "websocket url")
	user = flag.String("user", "", "username for accessing device")
	pass = flag.String("pass", "", "password for accessing device")
	dev = flag.String("device", "", "device")
	port = flag.Int("port", 0, "port")
	test = flag.Bool("test", false, "test")
	checkstate = flag.Bool("checkstate", false, "periodically check state of device")
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

	// Marker -> to see that program is working
	go func() {
		for {
			log.Println("--MARK--")
			time.Sleep(1 * time.Hour)
		}
	}()

	ctrlCHandler()

	for {
		startWebsocket(pin)

		//terminate current ExitHdl
		exitHdl <- false
	}

}

func startWebsocket(pin rpio.Pin) {
	log.Println("Connecting to " + *url)

	config, _ := websocket.NewConfig(*url, "http://localhost/")
	config.Header.Add("Authorization", "Basic "+basicAuth(*user, *pass))
	ws, err := websocket.DialConfig(config)
	if err != nil {
		log.Println(err)
		return
	}

	go exitHandler(ws, dev)

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
	currentState := getState(pin)
	transmitState(currentState, ws, dev)

	// // loop forever and read every second current state
	// if *checkstate {
	// 	for {
	// 		newstate := getState(pin)
	// 		log.Print("State from switch: ")
	// 		log.Println(newstate)
	// 		if newstate != currentState {
	// 			log.Println("NEW STATE FROM SWITCH: " + newstate)
	// 			transmitState(newstate, ws, dev)
	// 			currentState = newstate
	// 		}
	// 		time.Sleep(1 * time.Second)
	// 	}
	}

	// wait for end Handler
	<- endWS

}

func exitErr(cond bool, err string) {
	if cond {
		flag.PrintDefaults()
		log.Fatal(err)
	}
}

func exitHandler(ws *websocket.Conn, dev *string) {
	// wait for msg
	v := <-exitHdl

	// regular exit
	if v {
		sendToWebsocket(ws, &devcom.DevProto{
			Action: devcom.Disconnect,
			Device: devcom.Device{
				ID: *dev,
			},
		})
		ws.Close()
	}
}

func ctrlCHandler() {
	c := make(chan os.Signal, 1)
	signal.Reset(os.Interrupt)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		// regular exit
		exitHdl <- true

		log.Println("we received CTRL+C")
		done = true
		// exit hard if websocket answer hangs
		go func() {
			time.Sleep(5 * time.Second)
			os.Exit(0)
		}()

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

			switch incoming.PayLoad {
			case On:
				setState(pin, On)
			case Off:
				setState(pin, Off)
			}
		}
		// if done {
		// 	ws.Close()
		// 	log.Println("try to reconnect")
		// }
	}
	endWS <- true
}

func getState(pin rpio.Pin) string {
	x := pin.Read()

	log.Printf("got from device: %d\n", x)
	var state string
	if x == 1 {
		state = Off
	} else {
		state = On
	}
	return state
}

func setState(pin rpio.Pin, state string) {
	if state == On {
		pin.Low()
	}
	if state == Off {
		pin.High()
	}
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
	result := rpio.Pin(pin)
	result.Output()

	return result
}

func sendToWebsocket(ws *websocket.Conn, command *devcom.DevProto) {
	log.Printf("Sending %+v ...\n", command)
	err := websocket.JSON.Send(ws, command)
	if err != nil {
		log.Fatal(err)
	}
}
