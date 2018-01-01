package controllers

import (
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"schneidernet/smarthome/app"
	"schneidernet/smarthome/app/dao"
	"schneidernet/smarthome/app/models/alexa"
	"schneidernet/smarthome/app/routes"
	"strconv"
	"time"
)

/*
TODO:
- delete username from session
- IDEA: if first login -> set and display random device-password
*/

// Constants
const (
	REVELREDIRECT string = "REVEL_REDIRECT"
)

// Main Controller for WebApp
type Main struct {
	*revel.Controller
}

func init() {
	revel.InterceptMethod(Main.genericInterceptor, revel.BEFORE)
	revel.InterceptMethod(Main.checkUser, revel.BEFORE)
}

// Render Webpages
// *************************************************

// Index - Login Panel
func (c Main) Index() revel.Result {

	// if a valid session is present
	if _, validSession := c.Session["useroid"]; validSession {
		return c.Redirect(app.ContextRoot + routes.Main.Dashboard())
	}

	//Google Oauth2 Link
	// Es wird ein State über die URL durchgeschleift und später geprüft, ob dieser State noch da ist.
	state, url := getGoogleStateAndURL()
	c.ViewArgs["google_url"] = url
	c.Session[googleState] = state

	// Render LoginPanel
	return c.Render()
}

// Dashboard - Main-Page after Login
func (c Main) Dashboard() revel.Result {
	c.ViewArgs["devices"] = dao.GetAllDevices(c.getCurrentUserID())
	return c.Render()
}

// Settings - display Settings-Page
func (c Main) Settings() revel.Result {
	user := dao.GetUserWithID(c.getCurrentUserID())
	c.ViewArgs["muser"] = *user
	return c.Render()
}

// UserList - display UserList Page
func (c Main) UserList() revel.Result {
	users := dao.GetAllUsers()
	c.ViewArgs["users"] = *users
	return c.Render()
}

// Oauth - render Oauth-Page
func (c Main) Oauth2() revel.Result {
	c.ViewArgs["action"] = app.ContextRoot + "/oauth2/auth?" + c.Params.Query.Encode()
	for _, s := range []string{"redirect_uri", "scope", "client_id"} {
		c.ViewArgs[s] = c.Params.Query.Get(s)
	}
	return c.Render()
}

// DeviceList - List all Devices
func (c Main) DeviceList() revel.Result {
	devices := dao.GetAllDevices(c.getCurrentUserID())
	c.ViewArgs["devices"] = devices
	return c.Render()
}

// DeviceNew - Create Device Page
func (c Main) DeviceNew() revel.Result {
	i8nNames := []string{}
	for i := 0; i < alexa.DeviceTypeNum; i++ {
		i8nNames = append(i8nNames, c.Message("alexa.devicetype."+strconv.Itoa(i)))
	}
	c.ViewArgs["devicetype"] = i8nNames
	return c.Render()
}

// CreateDevicePassword ...
func (c Main) CreateDevicePassword() revel.Result {
	user := dao.GetUserWithID(c.getCurrentUserID())
	randompassword := "hallo"
	c.ViewArgs["pass"] = randompassword
	user.DevicePassword, _ = bcrypt.GenerateFromPassword([]byte(randompassword), bcrypt.DefaultCost)
	dao.SaveUser(user)

	return c.Render()
}

// Actions
// ***********************************

// Login processes a Login
func (c Main) Login(username, password string, remember bool) revel.Result {
	dbUsr := dao.GetUser(username)
	if dbUsr == nil {
		c.Flash.Error("User unbekannt")
		return c.Redirect(app.ContextRoot + routes.Main.Index())
	}

	err := bcrypt.CompareHashAndPassword(dbUsr.Password, []byte(password))
	if err != nil {
		c.Flash.Error("Password false")
		return c.Redirect(app.ContextRoot + routes.Main.Index())
	}

	// Set Session
	c.setSession(strconv.Itoa(int(dbUsr.ID)), dbUsr.UserID)

	return c.Redirect(c.getSuccessfulLoginRedirect())
}

// Logout API
func (c Main) Logout() revel.Result {
	delete(c.Session, "useroid")
	delete(c.Session, "userid")
	return c.Redirect(app.ContextRoot + routes.Main.Index())
}

// UpdateUser updates a User
func (c Main) UpdateUser(user dao.User) revel.Result {
	dbUser := dao.GetUserWithID(c.getCurrentUserID())
	dbUser.Name = user.Name
	dbUser.UserID = c.Session["userid"]
	dao.SaveUser(dbUser)
	return c.Redirect(app.ContextRoot + routes.Main.Dashboard())
}

// UserDel deletes a User
func (c Main) UserDel(id string) revel.Result {
	uid, _ := strconv.Atoi(id)
	user := dao.GetUserWithID(uint(uid))
	dao.DeleteUser(user)
	return c.Redirect(app.ContextRoot + routes.Main.Dashboard())
}

// CreateDevice creates a Device
func (c Main) CreateDevice(device dao.Device) revel.Result {
	device.UserID = c.getCurrentUserID()
	dao.CreateDevice(&device)
	return c.Redirect(app.ContextRoot + routes.Main.Dashboard())
}

// DeleteDevice deletes a Device
func (c Main) DeleteDevice(id string) revel.Result {
	device := dao.FindDeviceByID(c.getCurrentUserID(), id)
	//TODO: Popup Question
	//TODO: what if device do net exists

	dao.DeleteDevice(device)
	return c.Redirect(app.ContextRoot + routes.Main.DeviceList())
}

func (c Main) getCurrentUserID() uint {
	oid, _ := strconv.Atoi(c.Session["useroid"])
	return uint(oid)
}

// Interception Method gets called on every /main/ Request
func (c Main) genericInterceptor() revel.Result {
	// Set app.ContextRoot if we ar behind a rewritng Proxy
	c.ViewArgs["contextRoot"] = app.ContextRoot
	c.ViewArgs["websocketHost"] = app.WebSocketHost

	return nil
}

// Interception Method gets called on every /main/ Request
func (c Main) checkUser() revel.Result {

	// diese Seiten benötigen kein Login
	if c.Action == "Main.Index" ||
		c.Action == "Main.Login" ||
		c.Action == "Main.OAuth2CallBackGoogle" {
		return nil
	}

	// keine useroid -> zurück zur Loginseite
	if c.Session["useroid"] == "" {

		// if the Call is for the Websocket, request must have a basic auth header
		if c.Action == "Main.DeviceFeed" {
			return c.checkWebsocketBasicAuth()
		}

		c.Flash.Error("not logged in")

		// redirect back to intensional site
		c.SetCookie(&http.Cookie{
			Name:    REVELREDIRECT,
			Path:    "/",
			Value:   app.ContextRoot + c.Request.GetRequestURI(),
			Expires: time.Now().Add(time.Duration(5) * time.Minute),
		})

		return c.Redirect(app.ContextRoot + routes.Main.Index())
	}

	c.ViewArgs["user"] = c.Session["userid"]
	return nil
}

// redirect to the site the user want's initially (and would be intercepted by login page)
func (c Main) getSuccessfulLoginRedirect() string {

	// If User wants initially to an other site
	if cookie, err := c.Request.Cookie(REVELREDIRECT); err == nil {
		target := cookie.GetValue()

		// delete cookie
		c.SetCookie(&http.Cookie{
			Name:    REVELREDIRECT,
			Expires: time.Now().Add(time.Duration(-5) * time.Minute),
		})
		return target
	}

	// Default after Login
	return app.ContextRoot + routes.Main.Dashboard()
}

// Set the Session Info
func (c Main) setSession(useroid, userid string) {
	c.Session["useroid"] = useroid
	c.Session["userid"] = userid

}
