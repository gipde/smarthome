package controllers

import (
	"crypto/rand"
	"encoding/json"
	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/clientcredentials"
	"io/ioutil"
	mathrand "math/rand"
	"net/http"
	"net/url"
	"os"
	"schneidernet/smarthome/app"
	"schneidernet/smarthome/app/dao"
	"strings"
	"time"
)

type OAuthStorageAdapter struct{}

var (
	// Clients which will be allowed to connect to our Oauth2 Service
	clients         map[string]*fosite.DefaultClient
	oauth2Provider  fosite.OAuth2Provider
	smartHomeClient clientcredentials.Config
)

func init() {
	revel.AppLog.Debug("Init")
	revel.OnAppStart(initOauth2)
	revel.OnAppStart(installHandlers)

}

func initOauth2() {
	var clientconfig string
	clientconfig, ok := revel.Config.String("oauth.clients")
	if !ok {
		os.Exit(1)
	}
	file, err := ioutil.ReadFile(revel.BasePath + "/" + clientconfig)
	if err != nil {
		revel.AppLog.Errorf("File error: %v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(file, &clients)

	// Sepcial Case SmartHomeServer
	// this client will be used to query the introspection endpoint
	mathrand.Seed(time.Now().UnixNano())
	serverpass := make([]byte, 15)
	rand.Read(serverpass)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(serverpass), bcrypt.DefaultCost)

	smartHomeClient = clientcredentials.Config{
		ClientID:     "SmartHomeServer",
		Scopes:       []string{"devices"},
		ClientSecret: string(serverpass),
		TokenURL:     "http://localhost:9000/oauth2/token",
	}
	clients["SmartHomeServer"].Secret = hashedPassword

	initProvider()

	// Oauth2 Tokens
	go tokenCleaner()

}

func initProvider() {
	// Init Oauth2provider
	config := compose.Config{
		// AuthorizeCodeLifespan: time.Minute * 60,
		// AccessTokenLifespan:   time.Minute * 60,
	}
	strg := OAuthStorageAdapter{}

	oauthPass, _ := revel.Config.String("oauth.signingsecret")
	strat := compose.CommonStrategy{
		CoreStrategy: compose.NewOAuth2HMACStrategy(&config, []byte(oauthPass)),
	}

	oauth2Provider = compose.Compose(
		&config,
		strg,
		strat,
		nil,

		compose.OAuth2AuthorizeExplicitFactory,
		compose.OAuth2AuthorizeImplicitFactory,
		compose.OAuth2ClientCredentialsGrantFactory,
		compose.OAuth2RefreshTokenGrantFactory,
		compose.OAuth2ResourceOwnerPasswordCredentialsFactory,

		compose.OAuth2TokenRevocationFactory,
		compose.OAuth2TokenIntrospectionFactory,
	)
}

func newSession(user string) *fosite.DefaultSession {
	return &fosite.DefaultSession{
		ExpiresAt: map[fosite.TokenType]time.Time{
			fosite.AccessToken:   time.Now().UTC().Add(time.Hour),
			fosite.AuthorizeCode: time.Now().UTC().Add(time.Hour),
		},
		Username: user,
	}
}

func AuthorizeHandlerFunc(rw http.ResponseWriter, req *http.Request) {

	ctx := fosite.NewContext()

	ar, err := oauth2Provider.NewAuthorizeRequest(ctx, req)
	if err != nil {
		revel.AppLog.Errorf("Error occurred in NewAuthorizeRequest: %s", err)
		oauth2Provider.WriteAuthorizeError(rw, ar, err)
		return
	}

	// check if permission Accepted
	revel.AppLog.Debug("url", "permission", req.FormValue("permissionAccepted"))
	// Check a valid Revel-Session
	// Session is encrypted, so we can trust
	cook, err := req.Cookie("REVEL_SESSION")

	if req.FormValue("permissionAccepted") == "true" && err == nil && cook != nil {
		createResponse(cook, rw, ctx, ar, req)
		return
	}

	// Append Parameters to Login-Page for further Redirect back to here
	pars := req.Form.Encode()

	// No Valid-Session -> Redirect to Resource-Server
	http.Redirect(rw, req, app.PublicHost+app.ContextRoot+"/Main/Oauth2?"+pars, 302)
}

func createResponse(cook *http.Cookie, rw http.ResponseWriter, ctx context.Context, ar fosite.AuthorizeRequester, req *http.Request) {

	session := revel.GetSessionFromCookie(revel.GoCookie(*cook))
	revel.AppLog.Debug("session", "cookie", session.Cookie())

	if user, ok := session["userid"]; ok {

		//Grant every requested Scope
		for _, scope := range ar.GetRequestedScopes() {
			ar.GrantScope(scope)
		}

		createAuthorizeResponse(ctx, ar, rw, user)
		return
	}
	// what if no userid ?? FIX IT

}

func createAuthorizeResponse(ctx context.Context, ar fosite.AuthorizeRequester, rw http.ResponseWriter, user string) {
	var mySessionData *fosite.DefaultSession
	mySessionData = &fosite.DefaultSession{Username: user}

	// Now we need to get a response. This is the place where the AuthorizeEndpointHandlers kick in and start processing the request.
	// NewAuthorizeResponse is capable of running multiple response type handlers which in turn enables this library
	// to support open id connect.
	response, err := oauth2Provider.NewAuthorizeResponse(ctx, ar, mySessionData)
	if err != nil {
		oauth2Provider.WriteAuthorizeError(rw, ar, err)
		return
	}

	// Awesome, now we redirect back to the client redirect uri and pass along an authorize code
	oauth2Provider.WriteAuthorizeResponse(rw, ar, response)
}

// Handler für alle Token Requests (authorize,revoke, ...)
func TokenHandlerFunc(rw http.ResponseWriter, req *http.Request) {

	ctx := fosite.NewContext()

	// Create an empty session object which will be passed to the request handlers
	mySessionData := newSession("")

	// This will create an access request object and iterate through the registered TokenEndpointHandlers to validate the request.
	accessRequest, err := oauth2Provider.NewAccessRequest(ctx, req, mySessionData)

	// Catch any errors, e.g.:
	// * unknown client
	// * invalid redirect
	// * ...
	if err != nil {
		revel.AppLog.Errorf("Error occurred in NewAccessRequest: %s", err)
		oauth2Provider.WriteAccessError(rw, accessRequest, err)
		return
	}

	// Next we create a response for the access request. Again, we iterate through the TokenEndpointHandlers
	// and aggregate the result in response.
	response, err := oauth2Provider.NewAccessResponse(ctx, accessRequest)
	if err != nil {
		revel.AppLog.Errorf("Error occurred in NewAccessResponse: %s", err)
		oauth2Provider.WriteAccessError(rw, accessRequest, err)
		return
	}

	// if it is a token exchange, we save the id and the refresh token
	if accessRequest.GetGrantTypes().Has("authorization_code") {

		user := accessRequest.GetSession().GetUsername()
		dbUser := dao.GetUser(user)
		if dbUser == nil {
			revel.AppLog.Errorf("Logged In User %s nof found", user)
			return
		}
		if dbUser.Authorizations == nil {
			dbUser.Authorizations = []dao.AuthorizeEntry{}
		}
		// Save Auth App and Refresh-Token in DB
		dbUser.Authorizations = append(dbUser.Authorizations, dao.AuthorizeEntry{
			AppID:        accessRequest.GetClient().GetID(),
			RefreshToken: response.GetExtra("refresh_token").(string),
		})
		dao.SaveUser(dbUser)

	}

	// All done, send the response.
	oauth2Provider.WriteAccessResponse(rw, accessRequest, response)
	// The client now has a valid access token
}

// IntrospectionHandlerFunc -  we check if Token is valid
func IntrospectionHandlerFunc(rw http.ResponseWriter, req *http.Request) {

	ctx := fosite.NewContext()

	mySessionData := newSession("")
	ir, err := oauth2Provider.NewIntrospectionRequest(ctx, req, mySessionData)
	if err != nil {
		revel.AppLog.Errorf("Error occurred in NewAuthorizeRequest: %+v", err)
		oauth2Provider.WriteIntrospectionError(rw, err)
		return
	}
	oauth2Provider.WriteIntrospectionResponse(rw, ir)
}

// RevokationHanlderFunc ...
func RevocationHandlerFunc(rw http.ResponseWriter, req *http.Request) {

	ctx := fosite.NewContext()

	err := oauth2Provider.NewRevocationRequest(ctx, req)
	oauth2Provider.WriteRevocationResponse(rw, err)
}

// CheckToken - validates the token
func CheckToken(token string) (active bool, username string) {
	// TODO: Maybe use API directly instead of using HTTP/Localhost

	if ok, _ := revel.Config.Bool("oauth.checktoken"); !ok {
		user, _ := revel.Config.String("user.admin")
		return true, user
	}

	// we build clientcredentials from servercredentials
	data := url.Values{"token": {token}}
	client := smartHomeClient.Client(context.Background())

	// Firtst Request is a request to /token with client-credentails
	// Second one is a post request with the original auth token
	result, err := client.PostForm(strings.Replace(smartHomeClient.TokenURL, "token", "introspect", -1), data)
	if err != nil {
		revel.AppLog.Errorf("Error %v", err)
		return false, ""
	}
	defer result.Body.Close()

	var introspection = struct {
		Active   bool   `json:"active"`
		Username string `json:"username"`
	}{}
	out, _ := ioutil.ReadAll(result.Body)
	json.Unmarshal(out, &introspection)

	return introspection.Active, introspection.Username
}

/*
Oauth2 Storage Implementation
*/

func storeToken(signature string, tokenType fosite.TokenType, request fosite.Requester) error {
	serialized, _ := json.Marshal(request)
	expiry := request.GetSession().GetExpiresAt(tokenType)
	tokenid := request.GetID()
	if tokenType == fosite.RefreshToken {
		// refresh token expires in 10 years
		expiry = time.Now().AddDate(10, 0, 0)
	}

	err := dao.SaveToken(signature, tokenid, tokenType, expiry, &serialized)
	return err
}

func (c OAuthStorageAdapter) GetClient(ctx context.Context, id string) (fosite.Client, error) {
	cl, ok := clients[id]
	if !ok {
		return nil, fosite.ErrNotFound
	}
	return cl, nil
}

func (c OAuthStorageAdapter) CreateAuthorizeCodeSession(ctx context.Context, code string, request fosite.Requester) (err error) {
	return storeToken(code, fosite.AuthorizeCode, request)
}

func (c OAuthStorageAdapter) GetAuthorizeCodeSession(ctx context.Context, code string, session fosite.Session) (request fosite.Requester, err error) {
	var result = fosite.NewAuthorizeRequest()
	result.Session = session
	data := dao.GetTokenBySignature(code)
	if data == nil {
		return nil, fosite.ErrNotFound
	}
	json.Unmarshal(*data, &result)
	return result, nil
}

func (c OAuthStorageAdapter) InvalidateAuthorizeCodeSession(ctx context.Context, code string) (err error) {
	//TODO store active bool with token
	return nil
}

func (c OAuthStorageAdapter) DeleteAuthorizeCodeSession(ctx context.Context, code string) (err error) {
	err = dao.DeleteToken(code)
	return nil
}

func (c OAuthStorageAdapter) CreateAccessTokenSession(ctx context.Context, signature string, request fosite.Requester) (err error) {
	return storeToken(signature, fosite.AccessToken, request)
}

func (c OAuthStorageAdapter) GetAccessTokenSession(ctx context.Context, signature string, session fosite.Session) (request fosite.Requester, err error) {
	var result = fosite.NewAccessRequest(session)
	data := dao.GetTokenBySignature(signature)
	if data != nil {
		json.Unmarshal(*data, &result)
		if data == nil {
			return nil, fosite.ErrNotFound
		}
		return result, nil
	}
	return nil, fosite.ErrTokenSignatureMismatch
}

func (c OAuthStorageAdapter) DeleteAccessTokenSession(ctx context.Context, signature string) (err error) {
	return c.DeleteAuthorizeCodeSession(ctx, signature)
}

func (c OAuthStorageAdapter) CreateRefreshTokenSession(ctx context.Context, signature string, request fosite.Requester) (err error) {
	return storeToken(signature, fosite.RefreshToken, request)
}

func (c OAuthStorageAdapter) GetRefreshTokenSession(ctx context.Context, signature string, session fosite.Session) (request fosite.Requester, err error) {
	return c.GetAccessTokenSession(ctx, signature, session)
}

func (c OAuthStorageAdapter) DeleteRefreshTokenSession(ctx context.Context, signature string) (err error) {
	return c.DeleteAuthorizeCodeSession(ctx, signature)
}

func (c OAuthStorageAdapter) RevokeRefreshToken(ctx context.Context, requestID string) error {
	dao.DeleteTokenByTokenID(requestID, fosite.RefreshToken)
	return nil
}

func (c OAuthStorageAdapter) RevokeAccessToken(ctx context.Context, requestID string) error {
	dao.DeleteTokenByTokenID(requestID, fosite.AccessToken)
	return nil
}

// TODO: not implemented
func (c OAuthStorageAdapter) Authenticate(ctx context.Context, name string, secret string) error {
	revel.AppLog.Infof("Authenticate: %+v", ctx)
	return nil
}

func installHandlers() {
	revel.AddInitEventHandler(func(event int, _ interface{}) (r int) {
		if event == revel.ENGINE_STARTED {

			srvHandler := &revel.CurrentEngine.(*revel.GoHttpServer).Server.Handler
			revelHandler := *srvHandler

			serveMux := http.NewServeMux()
			serveMux.Handle("/", revelHandler) // the old handler
			serveMux.Handle("/oauth2/auth",
				http.HandlerFunc(AuthorizeHandlerFunc)) // authentication handler
			serveMux.Handle("/oauth2/token",
				http.HandlerFunc(TokenHandlerFunc)) // token handler (exchange,refresh,client-credentials)
			serveMux.Handle("/oauth2/introspect",
				http.HandlerFunc(IntrospectionHandlerFunc)) // introsepct token
			serveMux.Handle("/oauth2/revoke",
				http.HandlerFunc(RevocationHandlerFunc)) // revoke token
			*srvHandler = serveMux
		}
		return
	})
}

// Token Cleaner
func tokenCleaner() {
	for {
		dao.CleanExpiredTokens()
		time.Sleep(time.Minute)
	}
}
