package controllers

import (
	// "database/sql"
	//	"strings"

	// "fmt"

	"github.com/revel/revel"
	// "schneidernet/smarthome/app"
	"schneidernet/smarthome/app/routes"
	"strconv"
)

type Main struct {
	*revel.Controller
}

func (c Main) Index() revel.Result {
	c.Log.Info("Starte Loginseite")
	return c.Render()
}

func (c Main) Dashboard() revel.Result {
	if c.Session["key"] != "hello" {
		c.Flash.Error("not logged in")
		return c.Redirect(routes.Main.Index())
	}

	printResult(doSQL("SELECT * from tbl1;"))
	printResult(doSQL("SELECT name FROM sqlite_master WHERE type='table';"))

	return c.Render()
}

// REST Api
func (c Main) Login(username, password string, remember bool) revel.Result {

	c.Log.Info("User: " + username)
	c.Log.Info("pass: " + password)
	c.Log.Info("remember: " + strconv.FormatBool(remember))
	if username != "hello" {
		c.Flash.Error("Login failed")
		return c.Redirect(routes.Main.Index())
	}
	c.Session["key"] = "hello"

	return c.Redirect(routes.Main.Dashboard())

}

// Logout API
func (c Main) Logout() revel.Result {
	c.Session["key"] = ""
	return c.Redirect(routes.Main.Index())
}
