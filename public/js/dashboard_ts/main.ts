/*
embed in Dashboard.html
<script src="{{.contextRoot}}/public/js/dashboard_ts/main.js"></script> 
*/

declare var ReconnectingWebSocket: any
declare var $: any

const xpath = ""
const wspath = "ws://localhost:9000" + xpath
enum DeviceType { Socket = 0, Switch = 1, Bulb = 2, Thermostat = 3 }
const PingInterval:number=10

interface Device {
    id: string;
    connected?: boolean;
    devicetype?: DeviceType;
}
interface DevProto {
    action: string;
    device?: Device;
    payload?: string;
}




function send(action: string ,id:string): void {
    socket.send(JSON.stringify({
        action: action,
        device: { id: id}
    } as DevProto))
}

function setDevHdl(): void {
    $(".circle").each(function () { send("RequestState",$(this)[0].id) })
    $(".circle").click(function () { send("FlipState",$(this)[0].id) })
}

function drawSwitchable(incoming: DevProto, selector: any, part: string): void {
    if (incoming.payload == "ON") {
        selector.html("<img src=/public/img/" + part + "_on.svg width=30 height=30/>")
        selector.addClass("switchedOn")
    } else {
        selector.html("<img src=/public/img/" + part + "_off.svg width=30 height=30/>")
        selector.removeClass("switchedOn")
    }
}

function hdlIncoming(event: any): void {
    let data: DevProto = JSON.parse(event.data)
    let sel = $("#" + data.device.id)
    switch (data.action) {
        case "StateResponse":
        case "StateUpdate": {
            if (data.device.connected) {
                sel.addClass("connected")
            } else {
                sel.removeClass("connected")
            }
            switch (data.device.devicetype) {
                case DeviceType.Socket: {
                    drawSwitchable(data, sel, "socket")
                    break
                }
                case DeviceType.Switch: {
                    drawSwitchable(data, sel, "switch")
                    break
                }
                case DeviceType.Bulb: {
                    drawSwitchable(data, sel, "bulb")
                    break
                }
                case DeviceType.Thermostat: {
                    sel.html("<img src=/public/img/thermo.svg width=30 height=30/>" + data.payload)
                    break
                }
            }
            break;
        }
    }
}

function checkConnection(): number {
    let count=PingInterval
    return setInterval(function () {
        let color = socket.readyState == 1 ? "lightgreen" : "red"
        $("#wsstate").html('<svg height=25 width=25><circle cx=14 cy=18 r=5 stroke=black stroke-width=1 fill=' + color + '/></svg>')
    
        // ping
        count--
        if (count==0) {
            socket.send(JSON.stringify(<DevProto>{action:"Ping"}))
            count=PingInterval
        }
    }, 1000)
}

// main 

let socket = new ReconnectingWebSocket(wspath + '/Main/DeviceFeed');
let timer: number;

socket.onmessage = hdlIncoming
socket.onclose = function () {
    console.log("Websocket will be closed");
};

$(document).ready(function () {
    if (socket.readyState == 1) {
        setDevHdl()
    } else {
        socket.onopen = function () { setDevHdl() }
    }
    timer = checkConnection()
})

$(window).unload(function () {
    socket.close()
    console.log("We leave the page")
    window.clearInterval(timer)
});
