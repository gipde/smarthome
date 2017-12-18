package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"net/http"
	"os"
	"schneidernet/smarthome/app"
	"schneidernet/smarthome/app/dao"
	"schneidernet/smarthome/app/models"
	"schneidernet/smarthome/app/routes"
	"strconv"
	"strings"
	"time"
)

const (
	REVELREDIRECT string = "REVEL_REDIRECT"
)

type Main struct {
	*revel.Controller
}

func init() {
	revel.InterceptFunc(checkUser, revel.BEFORE, &Main{})
	revel.OnAppStart(initGoogleOauth2)
}

func (c Main) getCurrentUserID() uint {
	oid, _ := strconv.Atoi(c.Session["useroid"])
	return uint(oid)
}

func (c Main) Settings() revel.Result {
	user := dao.GetUserWithID(c.getCurrentUserID())
	c.ViewArgs["muser"] = *user
	c.Log.Infof("SettingsFlash: %+v", c.Flash.Data)
	return c.Render()
}

func (c Main) UserList() revel.Result {
	users := dao.GetAllUsers()
	c.ViewArgs["users"] = *users
	return c.Render()
}

func (c Main) UserDel(id string) revel.Result {
	c.Log.Infof("Delete User %s", id)
	uid, _ := strconv.Atoi(id)
	user := dao.GetUserWithID(uint(uid))
	dao.DeleteUser(user)
	return c.Redirect(app.ContextRoot + routes.Main.Dashboard())
}

func (c Main) UpdateUser(user dao.User) revel.Result {
	dbUser := dao.GetUserWithID(c.getCurrentUserID())
	dbUser.Name = user.Name
	c.Log.Infof("Speichere User: %+v", dbUser)
	dao.SaveUser(dbUser)
	return c.Redirect(app.ContextRoot + routes.Main.Dashboard())
}

func checkUser(c *revel.Controller) revel.Result {
	// Set app.ContextRoot if we ar behind a rewritng Proxy
	c.ViewArgs["contextRoot"] = app.ContextRoot
	c.ViewArgs["websocketHost"] = app.WebSocketHost

	// diese Seiten benötigen kein Login
	if c.Action == "Main.Index" ||
		c.Action == "Main.Login" ||
		c.Action == "Main.OAuth2CallBackGoogle" {
		return nil
	}

	// wenn das DeviceWebsocket aufgerufen wird, erlaube auch BasicAuth
	if c.Action == "Main.DeviceFeed" {
		if auth := c.Request.Header.Get("Authorization"); auth != "" {
			// Split up the string to get just the data, then get the credentials
			username, password, err := getCredentials(strings.Split(auth, " ")[1])
			if err != nil {
				return c.RenderError(err)
			}
			// TODO: check user/pass
			dbUsr := dao.GetUser(username)
			if dbUsr != nil {
				err := bcrypt.CompareHashAndPassword(dbUsr.Password, []byte(password))
				if err == nil {
					c.Session["useroid"] = strconv.Itoa(int(dbUsr.ID))
					c.Session["userid"] = dbUsr.UserID
					c.Log.Infof("Setting User into Session %+v", c.Session)
					// c.setUserSession(dbUsr)
					return nil
				}
			}
			return c.RenderError(errors.New("401: Not authorized"))
		}
	}

	// keine useroid -> zurück zur Loginseite
	if c.Session["useroid"] == "" {
		c.Flash.Error("not logged in")

		// redirect back to intensional site
		c.Log.Infof("I store a Cookie - you would like to open %s", app.ContextRoot+c.Request.GetRequestURI())
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

func (c Main) Oauth() revel.Result {
	c.ViewArgs["action"] = app.ContextRoot + "/oauth2/auth?" + c.Params.Query.Encode()
	for _, s := range []string{"redirect_uri", "scope", "client_id"} {
		c.ViewArgs[s] = c.Params.Query.Get(s)
	}
	return c.Render()
}

func (c Main) Index() revel.Result {

	// if a valid session is present
	if session, validSession := c.Session["useroid"]; validSession {
		c.Log.Infof("Session identified: %+v", session)
		return c.Redirect(app.ContextRoot + routes.Main.Dashboard())
	}

	//Google Oauth2 Link
	// Es wird ein State über die URL durchgeschleift und später geprüft, ob dieser State noch da ist.
	state := randToken()
	c.Session["state"] = state
	c.ViewArgs["google_url"] = conf.AuthCodeURL(state)

	// Render LoginPanel
	return c.Render()
}

func (c Main) Dashboard() revel.Result {
	c.ViewArgs["devices"] = dao.GetAllDevices(c.getCurrentUserID())
	return c.Render()
}

func CheckCredentials(username, password string) error {
	dbUsr := dao.GetUser(username)
	if username != dbUsr.UserID {
		return errors.New("user false")
	}
	err := bcrypt.CompareHashAndPassword(dbUsr.Password, []byte(password))
	if err != nil {
		return errors.New("password false")
	}
	return nil
}

func (c Main) getSuccessfulLoginRedirect() string {
	if cookie, err := c.Request.Cookie(REVELREDIRECT); err == nil {
		c.Log.Infof("we got a revelredirect-cookie %+v", cookie)
		target := cookie.GetValue()
		// cookie löschen
		c.SetCookie(&http.Cookie{
			Name:    REVELREDIRECT,
			Expires: time.Now().Add(time.Duration(-5) * time.Minute),
		})
		c.Log.Info("we redirect to " + target)
		return target
	}
	c.Log.Info("We redirect to Dashbaord")
	return app.ContextRoot + routes.Main.Dashboard()
}

// Function to generate a Hash from a Password
func (c Main) GetHash(password string) revel.Result {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	retval := struct{ Password []byte }{Password: hash}
	return c.RenderJSON(retval)

}

// REST Api
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
	c.setUserSession(dbUsr)

	return c.Redirect(c.getSuccessfulLoginRedirect())
}

func (c Main) setUserSession(dbUsr *dao.User) {
	c.Session["useroid"] = strconv.Itoa(int(dbUsr.ID))
	c.Session["userid"] = dbUsr.UserID
}

// Logout API
func (c Main) Logout() revel.Result {
	delete(c.Session, "useroid")
	delete(c.Session, "userid")
	return c.Redirect(app.ContextRoot + routes.Main.Index())
}

func (c Main) CreateDevice(device dao.Device) revel.Result {
	device.UserID = c.getCurrentUserID()
	dao.CreateDevice(&device)
	return c.Redirect(app.ContextRoot + routes.Main.Dashboard())
}

func (c Main) DeviceList() revel.Result {
	devices := dao.GetAllDevices(c.getCurrentUserID())
	for _, i := range *devices {
		c.Log.Info("Name: " + i.Name)
	}
	c.ViewArgs["devices"] = devices
	return c.Render()
}

func (c Main) DeleteDevice(id string) revel.Result {
	c.Log.Info("delete device: " + id + " ...")
	device := dao.FindDeviceByID(c.getCurrentUserID(), id)
	//TODO: Popup Question
	//TODO: what if device do net exists

	c.Log.Info(fmt.Sprintf("found device %v ", device))
	dao.DeleteDevice(device)
	return c.Redirect(app.ContextRoot + routes.Main.DeviceList())
}

func (c Main) DeviceNew() revel.Result {
	i8nNames := []string{}
	for i := 0; i < alexa.DeviceTypeNum; i++ {
		i8nNames = append(i8nNames, c.Message("alexa.devicetype."+strconv.Itoa(i)))
	}
	c.ViewArgs["devicetype"] = i8nNames
	return c.Render()
}

func (c Main) CreateDev() revel.Result {

	usr := dao.GetUserWithID(c.getCurrentUserID())
	if usr == nil {
		c.Log.Info("creating users")
		return nil
	}
	usr.Devices = []dao.Device{}

	for i := 0; i < 10; i++ {
		d := dao.Device{
			Name:        "Schalter",
			Description: "Schalter im FLur",
			Producer:    "WernerSchneiderNET",
			DeviceType:  2,
		}
		usr.Devices = append(usr.Devices, d)
	}
	dao.SaveUser(usr)
	return c.Redirect(app.ContextRoot + routes.Main.Dashboard())
}

func getSelList(c Main, idprefix string, count int) []string {
	list := make([]string, count)
	for i := 0; i < count; i++ {
		list[i] = c.Message(idprefix + "." + strconv.Itoa(i))
	}
	return list
}

var conf *oauth2.Config

type Credentials struct {
	Web struct {
		ClientID                string   `json:"client_id"`
		ProjectID               string   `json:"project_id"`
		AuthURI                 string   `json:"auth_uri"`
		TokenURI                string   `json:"token_uri"`
		AuthProviderX509CertURL string   `json:"auth_provider_x509_cert_url"`
		ClientSecret            string   `json:"client_secret"`
		RedirectURIS            []string `json:"redirect_uris"`
		JavascriptOrigins       []string `json:"javascript_origins"`
	}
}

type GUserInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamliyName    string `json:"family_name"`
	Profile       string `json:"profile"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
}

func initGoogleOauth2() {
	var c Credentials
	file, err := ioutil.ReadFile(revel.BasePath + "/conf/google_creds.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(file, &c)

	conf = &oauth2.Config{
		ClientID:     c.Web.ClientID,
		ClientSecret: c.Web.ClientSecret,
		RedirectURL:  app.PublicHost + app.ContextRoot + "/main/OAuth2CallBackGoogle",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			// You have to select your own scope from here ->
			// https://developers.google.com/identity/protocols/googlescopes#google_sign-in
		},
		Endpoint: google.Endpoint,
	}

}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func (c Main) OAuth2CallBackGoogle() revel.Result {

	retrievedState := c.Session["state"]
	if retrievedState != c.Params.Query.Get("state") {
		return c.RenderError(fmt.Errorf("Invalid session state: %s", retrievedState))
	}

	c.Log.Info("We have to fetch a new AccessToken")
	tok, err := conf.Exchange(oauth2.NoContext, c.Params.Query.Get("code"))
	if err != nil {
		return c.RenderError(err)
	}
	gToken, _ := json.Marshal(tok)
	c.Session["google"] = string(gToken)

	// if we have no GoogleAccess-Token we have to fetch one
	if _, ok := c.Session["google"]; !ok {
		c.Flash.Error("No Google Token Available")
		return c.Redirect(app.ContextRoot + routes.Main.Index())
	}
	c.Log.Info("We use our old AccessToken")
	// we use our previously fetched token

	// we check against Google
	client := conf.Client(oauth2.NoContext, tok)
	response, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return c.RenderError(err)
	}
	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)

	var userinfo GUserInfo
	json.Unmarshal(data, &userinfo)
	dbUsr := dao.GetUser(userinfo.Email)

	if dbUsr == nil {
		c.Log.Info("Kein Benutzer gefunden +%v", userinfo.Email)
		dbUsr = &dao.User{
			UserID: userinfo.Email,
			Name:   userinfo.Name,
		}
		dao.SaveUser(dbUsr)
	}
	c.setUserSession(dbUsr)
	c.Log.Info("Userinfo Mail: ", userinfo.Email)

	return c.Redirect(c.getSuccessfulLoginRedirect())

}

type DeviceCommand struct {
	Device     string
	Connected  bool
	Command    string
	State      string
	DeviceType int
}

type StateTopic struct {
	input    chan string
	consumer [](chan string)
}

var topics = make(map[uint]*StateTopic)

func register(uoid uint) (chan string, chan string) {
	if _, ok := topics[uoid]; !ok {
		// we create a new StateTopic
		topic := StateTopic{
			input:    make(chan string),
			consumer: [](chan string){},
		}
		topics[uoid] = &topic
		// and start a per user TopicHandler
		go topicHandler(&topic)
	}
	usertopic := topics[uoid]
	consumer := make(chan string)
	usertopic.consumer = append(usertopic.consumer, consumer)
	return usertopic.input, consumer
}

func unregister(uoid uint, consumer chan string) {
	revel.AppLog.Infof("Unregister Consumer %v for user %s -> %v", consumer, uoid, topics[uoid])
	usertopic := topics[uoid]

	for i, c := range usertopic.consumer {
		//TODO: check if equals is correct
		if c == consumer {
			// close consume and remove from list
			c <- "QUIT"
			close(c)
			usertopic.consumer = append(usertopic.consumer[:i], usertopic.consumer[i+1:]...)
			revel.AppLog.Infof("we found the consumer and close it List: %v", usertopic.consumer)
		}
	}

	// if this was the last consumer
	if len(usertopic.consumer) == 0 {
		// we can close the usertopic
		usertopic.input <- "QUIT"
		close(usertopic.input)
		delete(topics, uoid)
	}
	revel.AppLog.Infof("Unregister ready %v for user %s -> %v", consumer, uoid, topics[uoid])
}

func topicHandler(stateTopic *StateTopic) {
	go func() {
		for {
			msg := <-stateTopic.input
			revel.AppLog.Info("we got msg in topicHandler: " + msg)
			if msg == "QUIT" {
				revel.AppLog.Info("we leave Topic-Handler")
				break
			}
			// send to every consumer
			for _, consumer := range stateTopic.consumer {
				revel.AppLog.Infof("Sending to consumer %+v", consumer)
				consumer <- msg
				revel.AppLog.Infof("sent", consumer)
			}
		}
		revel.AppLog.Info("Leaving Topic Handler")
	}()
}

/*
Das Websocket empfängt Nachrichten von den Clients und sendet wieder
welche zurück.
Das Rendering ist grundsätzlich Aufgabe des Clients selbst

*/
func (c Main) DeviceFeed(ws revel.ServerWebSocket) revel.Result {
	c.Log.Infof("someone called a Websocket for ")

	usertopic, consumer := register(c.getCurrentUserID())

	//internal Receiver from StateTopic
	go func() {
		for {
			c.Log.Info("we are the internal consumer and wait for a msg")
			msg := <-consumer
			// after here, it is possible that WebSocketController is disabled
			revel.AppLog.Info("internal consumer got a msg: " + msg)
			if msg == "QUIT" {
				revel.AppLog.Info("Quiting Consumer")
				break
			}

			// send to Websocket
			var devState = DeviceCommand{}
			json.Unmarshal([]byte(msg), &devState)
			err := ws.MessageSendJSON(&devState)

			if err != nil {
				break
			}
		}
	}()

	//external Receiver from Websocket
	for {
		var msg string
		err := ws.MessageReceiveJSON(&msg)
		if err != nil {
			c.Log.Errorf("we got a error %v", err)
			break
		}

		var incoming DeviceCommand
		err = json.Unmarshal([]byte(msg), &incoming)
		if err != nil {
			c.Log.Errorf("Error in conversion %v", err)
		}
		c.Log.Infof("We got a message from Socket %s it's about device -> %+v", []byte(msg), incoming)

		dev := dao.FindDeviceByID(c.getCurrentUserID(), incoming.Device)
		c.Log.Infof("we found %+v", dev)

		incoming.DeviceType = dev.DeviceType

		switch incoming.Command {
		case "CLICK":
			incoming.Connected = dev.Connected
			if incoming.Connected {
				switch alexa.DeviceType(dev.DeviceType) {
				case alexa.DeviceLight,
					alexa.DeviceSocket,
					alexa.DeviceSwitch:

					c.Log.Infof("we have a thing to switch %d %s", dev.DeviceType, dev.State)
					if dev.State == "OFF" {
						dev.State = "ON"
					} else {
						dev.State = "OFF"
					}

					// if we have to send back
					dao.SaveDevice(dev)
					incoming.State = dev.State
					incoming.Command = "STATEUPDATE"

					j, _ := json.Marshal(&incoming)

					// send msg internally
					c.Log.Info("We send msg internally")
					usertopic <- string(j)
					c.Log.Info("sent")

				}
			}

		case "SETSTATE":
			if dev.Connected {
				dev.State = incoming.State
				dao.SaveDevice(dev)
				incoming.Command = "STATEUPDATE"
				incoming.Connected = dev.Connected
				j, _ := json.Marshal(&incoming)

				// send msg internally
				c.Log.Info("We send msg internally")
				usertopic <- string(j)
				c.Log.Info("sent")
			}

		case "GETSTATE":
			incoming.Command = "STATERESPONSE"
			incoming.State = dev.State
			incoming.Connected = dev.Connected
			err := ws.MessageSendJSON(&incoming)
			if err != nil {
				goto EXITLOOP
			}
		case "CONNECT":
			dev.Connected = incoming.Connected
			dao.SaveDevice(dev)
			incoming.Command = "STATEUPDATE"
			incoming.State = dev.State
			j, _ := json.Marshal(&incoming)

			// send msg internally
			c.Log.Info("We send msg internally")
			usertopic <- string(j)
			c.Log.Info("sent")

		case "DISCONNECT":
			dev.Connected = incoming.Connected
			dao.SaveDevice(dev)
			incoming.Command = "STATEUPDATE"
			incoming.State = dev.State
			j, _ := json.Marshal(&incoming)

			// send msg internally
			c.Log.Info("We send msg internally")
			usertopic <- string(j)
			c.Log.Info("sent")
		}

	}
EXITLOOP:

	unregister(c.getCurrentUserID(), consumer)
	c.Log.Info("we close the Websocket :(")
	return nil
}

func getCredentials(data string) (username, password string, err error) {
	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", "", err
	}
	strData := strings.Split(string(decodedData), ":")
	username = strData[0]
	password = strData[1]
	return
}
