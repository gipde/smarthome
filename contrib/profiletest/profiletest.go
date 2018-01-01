package main

import _ "net/http/pprof"
import "log"
import "net/http"
import "time"

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	for {
		time.Sleep(time.Minute * 1)
		println("Hello World")
	}
}
