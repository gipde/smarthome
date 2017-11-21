#!/usr/bin/env python

from BaseHTTPServer import BaseHTTPRequestHandler, HTTPServer
import SocketServer
from gpiozero import OutputDevice
from time import sleep

relais = OutputDevice(pin=4,active_high=False,initial_value=False)

def switch(state):
	if state:
		print("Switching on")
		relais.on()
	else:
		print("Switching off")
		relais.off()

def getState():
	return relais.value

class S(BaseHTTPRequestHandler):
    def _set_headers(self):
        self.end_headers()

    def do_GET(self):
	if self.path.lower() == "/on":
		switch(True)
	if self.path.lower() == "/off":
		switch(False)

	state="ON" if getState() else "OFF"
	json='{"state":"'+state+'"}'

        self.send_response(200)
        self.send_header('Content-type', 'application/json')
	self.send_header('content-length',len(json))
	self.end_headers()

	self.wfile.write(json)
	self.wfile.flush()

def run(server_class=HTTPServer, handler_class=S, port=80):
    server_address = ('', port)
    httpd = server_class(server_address, handler_class)
    print 'Starting httpd...'
    httpd.serve_forever()


if __name__ == "__main__":
    from sys import argv

    if len(argv) == 2:
        run(port=int(argv[1]))
    else:
        run()