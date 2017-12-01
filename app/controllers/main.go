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

	devices := dao.GetAllDevices()

	for _, i := range *devices {
		c.Log.Info("Name: " + i.Name)
	}
	c.ViewArgs["devices"] = devices

	c.ViewArgs["user"] = c.Session["user"]
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

	c.Session["logged_in"] = "true"
	c.Session["user"] = username
	return c.Redirect(routes.Main.Dashboard())

}

// Logout API
func (c Main) Logout() revel.Result {
	delete(c.Session, "logged_in")
	return c.Redirect(routes.Main.Index())
}

func (c Main) CreateDevice(device dao.Device) revel.Result {

	iface := c.Params.Form["iface"]
	var aInterfaces []dao.AlexaInterface
	for _, v := range iface {
		iv, _ := strconv.Atoi(v)
		aInterfaces = append(aInterfaces, dao.AlexaInterface{IFace: iv})
	}

	dCat := c.Params.Form["dcat"]
	var dCategories []dao.DisplayCategory
	for _, v := range dCat {
		iv, _ := strconv.Atoi(v)
		dCategories = append(dCategories, dao.DisplayCategory{DCat: iv})
	}

	device.AlexaInterfaces = aInterfaces
	device.DisplayCategories = dCategories
	

	dao.CreateDevice(&device)
	return c.Redirect(routes.Main.Dashboard())
}

func (c Main) DeviceList() revel.Result {
	devices := dao.GetAllDevices()
	for _, i := range *devices {
		c.Log.Info("Name: " + i.Name)
	}
	c.ViewArgs["devices"] = devices
	return c.Render()
}

func (c Main) DeleteDevice(id string) revel.Result {
	c.Log.Info("delete device: " + id + " ...")
	device := dao.FindDeviceByID(id)
	//TODO: Popup Question
	//TODO: what if device do net exists

	c.Log.Info(fmt.Sprintf("found device %v ", device))
	dao.DeleteDevice(device)
	return c.Redirect(routes.Main.DeviceList())
}

func (c Main) DeviceNew() revel.Result {
	c.ViewArgs["cats"] = getSelList(c, alexa.CategoryMessageID, alexa.CategoryNum)
	c.ViewArgs["interfaces"] = getSelList(c, alexa.CapabilityInterfaceMessageID, alexa.CapabilityInterfaceNums)
	return c.Render()
}

func getSelList(c Main, idprefix string, count int) []string {
	list := make([]string, count)
	for i := 0; i < count; i++ {
		list[i] = c.Message(idprefix + "." + strconv.Itoa(i))
	}
	return list
}
