package main

/*
include in Dashboard.html with

~/src/golang/src/schneidernet/smarthome/public/js/dashboard_go/main# gopherjs serve -vvvvv
<script src="http://localhost:8080/schneidernet/smarthome/public/js/dashboard_go/dashboard_go.js"></script>

gopherjs build schneidernet/smarthome/public/js/dashboard_go/main --localmap -w
<script src="{{.contextRoot}}/public/js/dashboard_go/main.js"></script>

*/

import (
	"encoding/json"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"schneidernet/smarthome/app/models/alexa"
	"schneidernet/smarthome/app/models/devcom"
)

// PingInterval in Seconds
const PingInterval = 300

var jq = jquery.NewJQuery

func getProto(action, id string) *string {
	res, _ := json.Marshal(&devcom.DevProto{
		Action: action,
		Device: devcom.Device{
			ID: id,
		},
	})
	retval := string(res)
	return &retval
}

func setDevHdl(ws *js.Object) {
	jq(".circle").Each(func(i int, iface interface{}) {
		ws.Call("send", getProto(devcom.RequestState, iface.(*js.Object).Get("id").String()))
	})
	jq(".circle").On(jquery.CLICK, func(e jquery.Event) {
		ws.Call("send", getProto(devcom.FlipState, e.CurrentTarget.Get("id").String()))
	})
}

func drawSwitchable(dev *devcom.DevProto, sel jquery.JQuery, part string) {
	if dev.PayLoad == alexa.ON {
		sel.SetHtml("<img src=/public/img/" + part + "_on.svg width=30 height=30/>")
		sel.AddClass("switchedOn")
	} else {
		sel.SetHtml("<img src=/public/img/" + part + "_off.svg width=30 height=30/>")
		sel.RemoveClass("switchedOn")
	}
}

func hdlIncoing(event *js.Object) {
	msg := event.Get("data").String()
	incoming := devcom.DevProto{}
	json.Unmarshal([]byte(msg), &incoming)
	sel := jq("#" + incoming.Device.ID)
	switch incoming.Action {
	case devcom.StateResponse:
		fallthrough
	case devcom.StateUpdate:
		if incoming.Device.Connected {
			sel.AddClass("connected")
		} else {
			sel.RemoveClass("connected")
		}
		switch incoming.Device.DeviceType {
		case alexa.DeviceSocket.ID():
			drawSwitchable(&incoming, sel, "socket")
		case alexa.DeviceSwitch.ID():
			drawSwitchable(&incoming, sel, "switch")
		case alexa.DeviceLight.ID():
			drawSwitchable(&incoming, sel, "bulb")
		case alexa.DeviceTemperatureSensor.ID():
			sel.SetHtml("<img src=/public/img/thermo.svg width=30 height=30/>" +
				incoming.PayLoad.(string))
		}
	}
}

func checkConnection(ws *js.Object) *js.Object {
	var pingInterval = PingInterval
	return js.Global.Call("setInterval", func() {
		var color string
		if ws.Get("readyState").Int() == 1 {
			color = "lightgreen"
		} else {
			color = "red"
		}
		jq("#wsstate").SetHtml(
			`<svg height=25 width=25>
                <circle cx=14 cy=18 r=5 stroke=black stroke-width=1 fill=` + color + `/>
			</svg>`)

		// ping
		pingInterval--
		if pingInterval == 0 {
			ping, _ := json.Marshal(&devcom.DevProto{Action: devcom.Ping})
			ws.Call("send", string(ping))
			pingInterval = PingInterval
		}

	}, 1000)
}

func main() {
	ws := js.Global.Get("ReconnectingWebSocket").New("ws://localhost:9000/Main/DeviceFeed")

	println("hello")
	ws.Set("onclose", func() {
		println("closing Websocket")
	})
	ws.Set("onmessage", hdlIncoing)

	jq("docment").Ready(func() {
		//if websocket is connected
		if ws.Get("readyState").Int() == 1 {
			setDevHdl(ws)
		} else {
			// set hdl on websocket connect
			ws.Set("onopen", func() { setDevHdl(ws) })
		}
	})

	conn := checkConnection(ws)

	window := js.Global.Get("window")
	jq(window).Call(jquery.UNLOAD, func() {
		ws.Call("close")
		window.Call("clearInterval", conn)
	})
}
