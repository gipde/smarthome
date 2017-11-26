package controllers

import (
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
	"schneidernet/smarthome/app"
	"schneidernet/smarthome/app/routes"
	"strconv"
)

type Main struct {
	*revel.Controller
}

func (c Main) Index() revel.Result {
	c.Log.Info("Starte Loginseite")
	c.Log.Info(c.Session["logged_in"])
	if c.Session["logged_in"] == "true" {
		return c.Redirect(routes.Main.Dashboard())
	}
	return c.Render()

}

func (c Main) Dashboard() revel.Result {
	if c.Session["logged_in"] != "true" {
		c.Flash.Error("not logged in")
		return c.Redirect(routes.Main.Index())
	}

	return c.Render()
}

// REST Api
func (c Main) Login(username, password string, remember bool) revel.Result {

	c.Log.Info("User: " + username)
	c.Log.Info("pass: " + password)
	c.Log.Info("remember: " + strconv.FormatBool(remember))

	var dbUsr app.User
	app.Db.First(&dbUsr, "id=?", username)

	if username != dbUsr.ID {
		c.Flash.Error("User not found")
		return c.Redirect(routes.Main.Index())
	}

	err := bcrypt.CompareHashAndPassword(dbUsr.Password, []byte(password))
	if err != nil {
		c.Flash.Error("Password false")
		return c.Redirect(routes.Main.Index())
	}

	c.Session["logged_in"] = "true"
	return c.Redirect(routes.Main.Dashboard())

}

// Logout API
func (c Main) Logout() revel.Result {
	delete(c.Session, "logged_in")
	return c.Redirect(routes.Main.Index())
}
