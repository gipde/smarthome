package controllers

import (
	"fmt"
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
	"schneidernet/smarthome/app/dao"
	"schneidernet/smarthome/app/models"
	"schneidernet/smarthome/app/routes"
	"strconv"
)

type Main struct {
	*revel.Controller
}

func init() {
	revel.InterceptFunc(checkUser, revel.BEFORE, &Main{})
}

func (c Main) getCurrentUser() string {
	return c.Session["useroid"]
}

func checkUser(c *revel.Controller) revel.Result {
	c.Log.Infof("Check User: %+v", c.Session)
	if c.Action != "Main.Index" && c.Action != "Main.Login" {
		if c.Session["useroid"] == "" {
			c.Flash.Error("not logged in")
			return c.Redirect(routes.Main.Index())
		}
	}
	c.ViewArgs["user"] = c.Session["userid"]
	return nil
}

func (c Main) Index() revel.Result {
	return c.Render()
}

func (c Main) Dashboard() revel.Result {
	c.ViewArgs["devices"] = dao.GetAllDevices(c.getCurrentUser())
	return c.Render()
}

// REST Api
func (c Main) Login(username, password string, remember bool) revel.Result {

	//TODO: Remember auswerten
	dbUsr := dao.GetUser(username)
	if username != dbUsr.UserID {
		c.Flash.Error("User unbekannt")
		return c.Redirect(routes.Main.Index())
	}

	err := bcrypt.CompareHashAndPassword(dbUsr.Password, []byte(password))
	if err != nil {
		c.Flash.Error("Password false")
		return c.Redirect(routes.Main.Index())
	}
	c.Session["useroid"] = strconv.Itoa(int(dbUsr.ID))
	c.Session["userid"] = dbUsr.UserID
	return c.Redirect(routes.Main.Dashboard())
}

// Logout API
func (c Main) Logout() revel.Result {
	delete(c.Session, "useroid")
	delete(c.Session, "userid")
	return c.Redirect(routes.Main.Index())
}

func (c Main) CreateDevice(device dao.Device) revel.Result {
	oid, _ := strconv.Atoi(c.getCurrentUser())
	device.UserID = oid
	dao.CreateDevice(&device)
	return c.Redirect(routes.Main.Dashboard())
}

func (c Main) DeviceList() revel.Result {
	devices := dao.GetAllDevices(c.getCurrentUser())
	for _, i := range *devices {
		c.Log.Info("Name: " + i.Name)
	}
	c.ViewArgs["devices"] = devices
	return c.Render()
}

func (c Main) DeleteDevice(id string) revel.Result {
	c.Log.Info("delete device: " + id + " ...")
	device := dao.FindDeviceByID(c.getCurrentUser(), id)
	//TODO: Popup Question
	//TODO: what if device do net exists

	c.Log.Info(fmt.Sprintf("found device %v ", device))
	dao.DeleteDevice(device)
	return c.Redirect(routes.Main.DeviceList())
}

func (c Main) DeviceNew() revel.Result {
	c.ViewArgs["devicetype"] = alexa.GetDeviceTypeNames()
	return c.Render()
}

func getSelList(c Main, idprefix string, count int) []string {
	list := make([]string, count)
	for i := 0; i < count; i++ {
		list[i] = c.Message(idprefix + "." + strconv.Itoa(i))
	}
	return list
}
