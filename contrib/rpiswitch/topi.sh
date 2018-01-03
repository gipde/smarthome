#!/bin/sh

env GOOS=linux GOARCH=arm GOARM=6 go build && scp rpiswitch pi@ds1820ws:/home/pi 
