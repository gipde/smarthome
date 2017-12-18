package controllers

import (
	"github.com/revel/revel"
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
