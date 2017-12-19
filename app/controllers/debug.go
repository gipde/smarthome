package controllers

import (
	"fmt"
	"github.com/revel/revel"
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
		html += "<td style='border: 1px solid black;'>" + t.Code + "</td>"
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
