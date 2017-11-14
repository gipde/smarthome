package controllers

import (
	"github.com/revel/revel"
)

type Oauth struct {
	*revel.Controller
}

type Token struct {
	AccesToken   string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

var redirectURL string
var state string
var clientID string

func (c Oauth) Authorize() revel.Result {
	clientID = c.Params.Query.Get("client_id")
	redirectURL = c.Params.Query.Get("redirect_uri")
	state = c.Params.Query.Get("state")
	scope := c.Params.Query.Get("scope")

	c.Log.Info(clientID)
	c.Log.Info(redirectURL)
	c.Log.Info(state)
	c.Log.Info(scope)
	return c.Redirect("/home/oauth/authenticate")
}

func (c Oauth) Token() revel.Result {
	return c.RenderJSON(Token{
		AccesToken:   "mF_9.B5f-4.1JqM",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: "tGzv3JOkF0XG5Qx2TlKWIA",
	})
}

func (c Oauth) Authenticate() revel.Result {
	return c.Render()
}

func (c Oauth) AuthorizeResponse() revel.Result {
	// Code lifetime max 10 min -> uuid
	code := "SplxlOBeZQQYbYS6WxSbIA"

	c.Response.ContentType = "application/x-www-form-urlencoded"
	return c.Redirect(redirectURL + "?code=" + code + "&state=" + state)
}



/*
Ablauf
1. Die Clientwebap ruft den Resource-Server mit clientid  usw. auf
2. der resource-server authentifiziert den user
	- im Schritt 1 erfolgt ein Redirect da der User noch nicht eingeloggt ist.
	- user muss Authorisierung bestätigen
3. der resource-server sendet einen redirect an die in 1 übergebene url hier enthalten ist der
   authorization code (z.B. http://localhost:9094/oauth2?code=4BFSDCFJP6SEIKAGE6ZSXA&state=xyz)
4. der autorization code wird durch einen access token getauscht
   hierbei wird clientid, clientsecret, granttype
5. der Server antwortet mit access-token
6. die cleintwebap kann mit dem access-token auf resourcen zugreifen
*/
