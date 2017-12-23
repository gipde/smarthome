package app

import (
	"bytes"
	"fmt"
	"github.com/revel/revel"
	"io/ioutil"
	"net/http"
)

// DoLogHHTPRequest logs a http Request
func DoLogHTTPRequest(r *http.Request, prefix string) {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("%v %v %v\n", r.Method, r.URL, r.Proto))
	buffer.WriteString(fmt.Sprintf("Host: %v\n", r.Host))
	for name, value := range r.Header {
		buffer.WriteString(fmt.Sprintf("Header: %s -> %s\n", name, value))
	}

	if r.Method == "POST" {
		r.ParseForm()
		buffer.WriteString("\n")
		buffer.WriteString(r.Form.Encode() + "\n")
	}

	bodyBuffer, _ := ioutil.ReadAll(r.Body)
	buffer.WriteString(fmt.Sprintf("Body: \n%s\n", bodyBuffer))

	revel.AppLog.Infof("HTTP Request: %s\n%s", prefix, buffer.String())

}
