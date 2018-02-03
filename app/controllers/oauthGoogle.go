package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/revel/revel"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"os"
	"schneidernet/smarthome/app"
	"schneidernet/smarthome/app/dao"
	gmodel "schneidernet/smarthome/app/models/google"
	"strconv"
)

const (
	googleState = "googleState"
)

var conf *oauth2.Config

func init() {
	revel.AppLog.Debug("Init")
	revel.OnAppStart(initGoogleOauth2)
}

// init Google Oauth2 Provider
func initGoogleOauth2() {

	c := readGoogleCredentials()

	conf = &oauth2.Config{
		ClientID:     c.Web.ClientID,
		ClientSecret: c.Web.ClientSecret,
		RedirectURL:  app.PublicHost + app.ContextRoot + "/main/OAuth2CallBackGoogle",
		Scopes: []string{
			// You have to select your own scope from here ->
			// https://developers.google.com/identity/protocols/googlescopes#google_sign-in

			// We use Userinfo.Email
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}

}

// Oauth2CallBackGoogle will be called, after we successfully authenticated against Google
func (c Main) OAuth2CallBackGoogle() revel.Result {
	// URL arguments
	// - state -> state we delivered to Google
	// - code -> authorization Code

	//check google state
	retrievedState := c.Session[googleState]
	if retrievedState != c.Params.Query.Get("state") {
		return c.RenderError(errors.New("Invalid session state"))
	}

	// We have to Fetch a new AccessToken from Google
	tok, err := conf.Exchange(oauth2.NoContext, c.Params.Query.Get("code"))
	if err != nil {
		return c.RenderError(err)
	}

	// we use our access-token and fetch mail info from  Google
	userinfo, err := getGoogleUserInfo(tok)
	if err != nil {
		return c.RenderError(err)
	}
	user := dao.GetOrCreateUser(userinfo.Email, userinfo.Name)

	// Set Session Info
	c.setSession(strconv.Itoa(int(user.ID)), user.UserID)

	return c.Redirect(c.getSuccessfulLoginRedirect())
}

// get Google Userinfo
func getGoogleUserInfo(tok *oauth2.Token) (*gmodel.GUserInfo, error) {
	client := conf.Client(oauth2.NoContext, tok)
	response, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	data, _ := ioutil.ReadAll(response.Body)

	var userinfo gmodel.GUserInfo
	json.Unmarshal(data, &userinfo)
	return &userinfo, nil
}

// create the URL for Login with Google
func getGoogleStateAndURL() (state string, url string) {
	state = app.RandToken(32)
	url = conf.AuthCodeURL(state)
	return state, url
}

// read Credentials we have from Google
func readGoogleCredentials() *gmodel.Credentials {
	var c gmodel.Credentials
	file, err := ioutil.ReadFile(revel.BasePath + "/conf/google_creds.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(file, &c)
	return &c
}
