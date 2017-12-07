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
	"schneidernet/smarthome/app/dao"
	"schneidernet/smarthome/app/models"
	"schneidernet/smarthome/app/routes"
	"strconv"
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

func (c Main) getCurrentUser() string {
	return c.Session["useroid"]
}

func (c Main) Settings() revel.Result {
	oid, _ := strconv.Atoi(c.getCurrentUser())
	user := dao.GetUserWithID(oid)
	c.ViewArgs["muser"] = *user
	c.Log.Infof("SettingsFlash: %+v", c.Flash.Data)
	return c.Render()
}

func (c Main) UpdateUser(user dao.User) revel.Result {
	oid, _ := strconv.Atoi(c.getCurrentUser())
	dbUser := dao.GetUserWithID(oid)
	dbUser.Name = user.Name
	c.Log.Infof("Speichere User: %+v", dbUser)
	dao.SaveUser(dbUser)
	return c.Redirect(routes.Main.Dashboard())
}

func checkUser(c *revel.Controller) revel.Result {
	c.Log.Infof("Check User: %+v in %+v", c.Session, c.Action)

	if _, ok := c.Params.Query["checkOauth2"]; ok {
		c.Log.Info("Checkin in oauth mode")
		c.Flash.Data["oauth"] = "true"
	}

	isin := func(literal string, list []string) bool {
		for _, v := range list {
			if v == literal {
				return true
			}
		}
		return false
	}
	if isin(c.Action, []string{
		"Main.Index",
		"Main.Login",
		"Main.OAuth2CallBackGoogle",
	}) {
		return nil
	}

	if c.Session["useroid"] == "" {
		c.Flash.Error("not logged in")

		new_cookie := &http.Cookie{
			Name:    REVELREDIRECT,
			Value:   c.Request.GetRequestURI(),
			Expires: time.Now().Add(time.Duration(5) * time.Minute),
		}
		c.SetCookie(new_cookie)
		return c.Redirect(routes.Main.Index())
	}

	c.ViewArgs["user"] = c.Session["userid"]
	return nil
}

func (c Main) Oauth() revel.Result {
	return c.Render()
}

func (c Main) Index() revel.Result {

	// when request comes for oauth -> LoginPanel

	if c.Flash.Data["oauth"] == "true" {
		// wenn bereits eine gültige Sitzung vorhanden ist, muss nur noch ein Bestätigungsdialog angezeigt werden

		c.ViewArgs["action"] = "/oauth2/auth"

		acc := map[string]string{}
		f := func(keys []string) {
			for _, v := range keys {
				acc[v] = c.Params.Get(v)
			}
		}

		f([]string{"client_id", "redirect_uri", "response_type", "scope", "state", "nonce"})
		c.ViewArgs["hidden"] = acc
		return c.Render()
	}

	// wenn gültie Session vorhanen, dann gleich weiter zum Dashboard
	if _, ok := c.Session["useroid"]; ok {
		return c.Redirect(routes.Main.Dashboard())
	}

	//Google Oauth2
	// Es wird ein State über die URL durchgeschleift und später geprüft, ob dieser State noch da ist.
	state := randToken()
	c.Session["state"] = state
	c.ViewArgs["google_url"] = conf.AuthCodeURL(state)

	c.ViewArgs["action"] = "/main/login"
	return c.Render()
}

func (c Main) Dashboard() revel.Result {
	c.Log.Infof("engin name: %+v", (*revel.GoEngine).Name)
	c.ViewArgs["devices"] = dao.GetAllDevices(c.getCurrentUser())
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

// REST Api
func (c Main) Login(username, password string, remember bool) revel.Result {

	target := routes.Main.Dashboard()

	if v, err := c.Request.Cookie(REVELREDIRECT); err == nil {
		target = v.GetValue()
		new_cookie := &http.Cookie{
			Name:    REVELREDIRECT,
			Value:   target,
			Expires: time.Now().Add(time.Duration(-5) * time.Minute),
		}
		c.SetCookie(new_cookie)
	}

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
	c.setUserSession(dbUsr)
	c.Log.Info("we redirect to " + target)
	return c.Redirect(target)
}

func (c Main) setUserSession(dbUsr *dao.User) {
	c.Session["useroid"] = strconv.Itoa(int(dbUsr.ID))
	c.Session["userid"] = dbUsr.UserID
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
	file, err := ioutil.ReadFile(revel.BasePath + "/secret/creds.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(file, &c)

	conf = &oauth2.Config{
		ClientID:     "my-client",
		ClientSecret: c.Web.ClientSecret,
		RedirectURL:  "http://localhost:9000/main/OAuth2CallBackGoogle",
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
		return c.Redirect(routes.Main.Index())
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

	return c.Redirect(routes.Main.Dashboard())

}
