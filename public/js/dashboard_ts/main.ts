import * as ReconnectingWebsocket from "reconnecting-websocket";

/*
embed in Dashboard.html
<script src="{{.contextRoot}}/public/js/dashboard_ts/main.js"></script>

install all deps
 > npm install

link global typescript
 > npm link typescript

run the build via webpack (see package.json)
 > npm run build

in Dev-Mode (with file watcher)
 > npm run watch

*/

const PingInterval: number = 300
const ON = "ON"
const SWITCHED_ON = "switchedOn"

enum DeviceType {
    Socket = 0,
    Switch = 1,
    Bulb = 2,
    Thermostat = 3
}

enum Action {
    Ping = "Ping",
    Pong = "Pong",
    StateResponse = "StateResponse",
    StateUpdate = "StateUpdate",
    RequestState = "RequestState",
    FlipState = "FlipState",
}

interface Device {
    id: string
    connected?: boolean
    devicetype?: DeviceType
}

interface DevProto {
    action: Action
    device?: Device
    payload?: any
}

class Dashboard {

    socket: ReconnectingWebsocket
    timer: number
    contextRoot:string

    drawSwitchable(incoming: DevProto, selector: JQuery, svg_part: string): void {
        if (incoming.payload == ON) {
            selector.html("<img src="+this.contextRoot+"/public/img/" + svg_part + "_on.svg width=30 height=30/>")
            selector.addClass(SWITCHED_ON)
        } else {
            selector.html("<img src="+this.contextRoot+"/public/img/" + svg_part + "_off.svg width=30 height=30/>")
            selector.removeClass(SWITCHED_ON)
        }
    }

    getIncomingHdl() {
        return (event: any) => {
            let data: DevProto = JSON.parse(event.data)
            let sel = $("#" + data.device.id)
            switch (data.action) {
                case Action.StateResponse:
                case Action.StateUpdate: {
                    if (data.device.connected) {
                        sel.addClass("connected")
                    } else {
                        sel.removeClass("connected")
                    }
                    switch (data.device.devicetype) {
                        case DeviceType.Socket: {
                            this.drawSwitchable(data, sel, "socket")
                            break
                        }
                        case DeviceType.Switch: {
                            this.drawSwitchable(data, sel, "switch")
                            break
                        }
                        case DeviceType.Bulb: {
                            this.drawSwitchable(data, sel, "bulb")
                            break
                        }
                        case DeviceType.Thermostat: {
                            sel.html("<img src="+this.contextRoot+"/public/img/thermo.svg width=30 height=30/>" + data.payload)
                            break
                        }
                    }
                    break
                }
                case Action.Ping: {
                    this.socket.send(JSON.stringify({
                        action: Action.Pong
                    } as DevProto))
                    break
                }

                case Action.Pong: {
                    $("#wsstate").attr("title",data.payload.ID)
                }
            }
        }
    }

    checkConnection(): number {
        let count = 0
        return setInterval(() => {
            // ping
            if (count == 0) {
                this.socket.send(JSON.stringify(<DevProto>{ action: Action.Ping }))
                count = PingInterval
            } else {
                count--
            }

            let color = this.socket.readyState == 1 ? "lightgreen" : "red"
            $("#wsstate").html('<svg height=25 width=25><circle cx=14 cy=18 r=5 stroke=black stroke-width=1 fill=' + color + '/></svg>')

        }, 1000)
    }

    send(action: Action, id?: string): void {
        this.socket.send(JSON.stringify({
            action: action,
            device: { id: id }
        } as DevProto))
    }

    setDevHdl(): void {
        let instance = this
        $(".circle").each(function (this: HTMLDivElement, _index: number, _elem: Element) {
            instance.send(Action.RequestState, $(this)[0].id)
        })
        $(".circle").click(function (this: HTMLDivElement, _eventObject: JQueryEventObject) {
            instance.send(Action.FlipState, $(this)[0].id)
        })
    }

    constructor() {

        // when Document ready rendered
        $(document).ready(() => {
            this.contextRoot=$("#wsstate").attr("contextroot")
            let url=$("#wsstate").attr("wshost") + this.contextRoot+ '/Main/DeviceFeed'
            this.socket = new ReconnectingWebsocket(url)
 
            // set Handler
            this.socket.onmessage = this.getIncomingHdl()
            this.socket.onclose = function () {
                console.log("Websocket will be closed")
            }

            if (this.socket.readyState == 1) {
                this.setDevHdl()
            } else {
                this.socket.onopen = () => { this.setDevHdl() }
            }
            // check the connection & ping server
            this.timer = this.checkConnection()
        })

        // Unload Window
        $(window).unload(() => {
            console.log("We leave the page")
            this.socket.close()
            window.clearInterval(this.timer)
        })
    }
}

new Dashboard()
