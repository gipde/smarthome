/*
closure compiler
java -jar $HOME/Downloads/closure-compiler-v20180101.jar --js dashboard.js --compilation_level SIMPLE --js_output_file out.js

embed in Dashboard.html
<script src="{{.contextRoot}}/public/js/reconnecting-websocket.js"></script>
<script src="{{.contextRoot}}/public/js/dashboard_js/out.js"></script>

*/

"use strict";
var xpath = "";
var wspath = "ws://localhost:9000" + xpath;
var socket = new ReconnectingWebSocket(wspath + '/Main/DeviceFeed');
var timer;


// send Click Command on every Click
var callbackhandler = function() {
    $(".circle").click(function() {
        var transmit = {
            "action": "FlipState",
            "device": {
                "devicetype": 0,
                "connected": true,
                "id": $(this)[0].id
            },
            "payload": ""
        };
        socket.send(JSON.stringify(transmit));
    });
};
// initially, we fetch the state of the device
var loadhandler = function() {
    $(".circle").each(function(index) {
        socket.send(JSON.stringify({
            "Action": "RequestState",
            "Device": {
                "ID": $(this)[0].id
            }
        }));
    });
};
$(document).ready(function() {
    //display connection state
    var interval = 100;
    timer = setInterval(function() {
        $("#wsstate").html(function() {
            var color;
            if (socket.readyState == 1)
                color = "lightgreen";
            else
                color = "red";
            return "<svg height=25 width=25><circle cx=14 cy=18 r=5 stroke=black stroke-width=1 fill=" + color + "/>\n                        </svg>";
        })

        //ping 
        interval--
        if (interval == 0) {
            interval = 100
            socket.send(JSON.stringify({ "Action": "Ping" }))
        }

    }, 1000);

    // Message received on the socket
    socket.onmessage = function(event) {
        var state = JSON.parse(event.data);
        console.log(state);
        var devid = state.device.id;
        switch (state.action) {
            case "StateUpdate": //fall through
            case "StateResponse":
                // Text
                $("#" + devid).html(state.payload);
                // State-Border
                if (state.device.connected) {
                    $("#" + devid).addClass("connected");
                } else {
                    $("#" + devid).removeClass("connected");
                }
                switch (state.device.devicetype) {
                    case 0:
                        //socket
                        console.log("we process for a socket a " + state.payload);
                        if (state.payload == "ON") {
                            $("#" + devid).html(function() {
                                return '<img src="' + xpath + '/public/img/socket_on.svg" width="30" height="30"/>';
                            });
                            $("#" + devid).addClass("switchedOn");
                        } else {
                            $("#" + devid).html(function() {
                                return '<img src="' + xpath + '/public/img/socket_off.svg" width="35" height="35"/>';
                            });
                            $("#" + devid).removeClass("switchedOn");
                        }
                        break;
                    case 1:
                        //switch
                        console.log("we process for a switch a " + state.payload);
                        if (state.payload == "ON") {
                            $("#" + devid).html(function() {
                                return '<img src="' + xpath + '/public/img/switch_on.svg" width="40" height="40"/>';
                            });
                            $("#" + devid).addClass("switchedOn");
                        } else {
                            $("#" + devid).html(function() {
                                return '<img src="' + xpath + '/public/img/switch_off.svg" width="40" height="40"/>';
                            });
                            $("#" + devid).removeClass("switchedOn");
                        }
                        break;
                    case 2:
                        // bulb
                        console.log("we process for a bulb " + state.payload);
                        if (state.payload == "ON") {
                            $("#" + devid).addClass("switchedOn");
                            $("#" + devid).html(function() {
                                return '<img src="' + xpath + '/public/img/bulb_on.svg" width="30" height="30"/>';
                            });
                        } else {
                            $("#" + devid).html(function() {
                                return '<img src="' + xpath + '/public/img/bulb_off.svg" width="30" height="30"/>';
                            });
                            $("#" + devid).removeClass("switchedOn");
                        }
                        break;
                    case 3:
                        // thermo
                        $("#" + devid).html(function() {
                            return '<img src="' + xpath + '/public/img/thermo.svg" width="25" height="25"/>' + state.payload;
                        });
                        break;
                }
                break;
        }
    };
    if (socket.readyState != 1)
        socket.onopen = function() {
            callbackhandler();
            loadhandler();
        };
    else {
        callbackhandler();
        loadhandler();
    }
    socket.onclose = function() {
        console.log("Websocket will be closed");
    };
    // wait until socket is connected
    // TODO: Nice UI State to display state of Websocket
    var connectionError = $("#connection-error");
    setTimeout(function() {
        if (socket.readyState != 1) {
            connectionError.html("error connecting websocket");
        }
    }, 1000); // ms
});

$(window).unload(function() {
    socket.close();
    console.log("We leave the page");
    window.clearInterval(timer);
});