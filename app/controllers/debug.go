package controllers

import (
	"fmt"
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
	"schneidernet/smarthome/app"
	"schneidernet/smarthome/app/dao"
	"strconv"
	"time"
)

type Debug struct {
	*revel.Controller
}

func (c Debug) ListTokens() revel.Result {
	html := "<table style='width:100%;border: 1px solid black;'>"
	// <tr><th>ID</th><th>Payload</th></tr>"
	tokens := dao.GetAllTokens()
	for _, t := range *tokens {
		html += "<tr>"
		html += "<td style='border: 1px solid black;'>" + strconv.Itoa(int(t.ID)) + "</td>"
		html += "<td style='border: 1px solid black;'>" + t.Signature + "</td>"
		html += "<td style='border: 1px solid black;'>" + t.TokenID + "</td>"
		html += "<td style='border: 1px solid black;'>" + string(t.TokenType) + "</td>"
		html += "<td style='border: 1px solid black;'>" + t.Expiry.Format(time.RFC3339) + "</td>"
		html += "<td style='border: 1px solid black;'>" + string(t.PayLoad) + "</td>"
		html += "</tr>"
	}
	return c.RenderHTML(html)
}

func (c Debug) CheckToken(token string) revel.Result {
	valid, user := app.CheckToken(token)
	return c.RenderText(fmt.Sprintf("active: %s\nuser: %s\n", valid, user))
}

// Function to generate a Hash from a Password
func (c Debug) GetHash(password string) revel.Result {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	retval := struct{ Password []byte }{Password: hash}
	return c.RenderJSON(retval)
}

// Log a Request
func (c Debug) LogRequest() revel.Result {
	DoLogRevelRequest(c.Request, "LoggingAPI Endpoint")
	return c.NotFound("but your Request is Logged :)")
}

func DoLogRevelRequest(req *revel.Request, prefix string) {
	originalHeader := req.Header.Server.(*revel.GoHeader)
	r := originalHeader.Source.(*revel.GoRequest).Original
	app.DoLogHTTPRequest(r, prefix)
}
