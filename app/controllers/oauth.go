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
