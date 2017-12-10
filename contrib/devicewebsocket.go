package main

import (
	"bufio"
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"os"
	"strings"
)

type Request struct {
	User     string
	Password string
	Device   string
	ReqType  string
	Payload  string
}

type DevState struct {
	Device    string
	State     string
	Connected bool
}

func doSwitch(ws *websocket.Conn, t int) error {
	req := Request{
		User:     "admin",
		Password: "admin",
		Device:   "1",
	}
	switch t {
	case 0:
		req.ReqType = "SWITCH_ON"
	case 1:
		req.ReqType = "SWITCH_OFF"
	case 2:
		req.ReqType = "CONNECT"
	case 3:
		req.ReqType = "DISCONNECT"
	}
	fmt.Printf("Send: %+v\n", req)
	return websocket.JSON.Send(ws, &req)
}

func main() {
	fmt.Println("Hellow Orld")

	origin := "http://localhost/"
	url := "ws://localhost:9000/devicewebsocket/devicesocket"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			var msg = DevState{}
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
		case "on":
			err = doSwitch(ws, 0)
		case "off":
			err = doSwitch(ws, 1)
		case "connect":
			err = doSwitch(ws, 2)
		case "disconnect":
			err = doSwitch(ws, 3)

		}

		if exit || err != nil {
			break
		}

	}

	ws.Close()

}
