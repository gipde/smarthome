package app

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/pkg/errors"
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/clientcredentials"
	"io/ioutil"
	mathrand "math/rand"
	"net/http"
	"net/url"
	"os"
	"schneidernet/smarthome/app/dao"
	"strings"
	"time"
)

type OAuthStorageAdapter struct{}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

var (
	// Clients which will be allowed to connect to our Oauth2 Service
	clients map[string]*fosite.DefaultClient
)

var (
	oauth2Provider  fosite.OAuth2Provider
	smartHomeClient clientcredentials.Config
)

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

	// Token Cleaner
	go tokenCleaner()

}

func initProvider() {
	// Init Oauth2provider
	config := compose.Config{
		AuthorizeCodeLifespan: time.Minute * 5,
		AccessTokenLifespan:   time.Minute * 10,
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

// Token Clenaer -> remove outdated-Tokens
// TODO: wird ein abgelaufener Token nicht automatisch entfernt ???
func tokenCleaner() {
	for {
		dao.CleanExpiredTokens()
		time.Sleep(time.Minute)
	}
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
	// revel.AppLog.Infof("new OAuth2 Authorize Request from %s", req.RemoteAddr)
	DoLogHTTPRequest(req, "Auth-Handler")
	ctx := fosite.NewContext()

	ar, err := oauth2Provider.NewAuthorizeRequest(ctx, req)
	if err != nil {
		revel.AppLog.Errorf("Error occurred in NewAuthorizeRequest: %s\nStack: \n%s", err, err.(stackTracer).StackTrace())
		oauth2Provider.WriteAuthorizeError(rw, ar, err)
		return
	}

	// Check a valid Revel-Session
	// Session is encrypted, so we can trust
	cook, err := req.Cookie("REVEL_SESSION")
	if err == nil && cook != nil {

		session := revel.GetSessionFromCookie(revel.GoCookie(*cook))

		if user, ok := session["userid"]; ok {

			//Grant every requested Scope
			for _, scope := range ar.GetRequestedScopes() {
				ar.GrantScope(scope)
			}

			createAuthorizeResponse(ctx, ar, rw, user)
			return
		}
	}

	// Append Parameters to Login-Page for further Redirect back to here
	pars := req.Form.Encode()

	// No Valid-Session -> Redirect to Resource-Server
	http.Redirect(rw, req, PublicHost+ContextRoot+"/Main/Oauth?"+pars, 302)
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
	// revel.AppLog.Infof("new OAuth2 Token Request from %s", req.RemoteAddr)
	DoLogHTTPRequest(req, "Auth-Handler")

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
		revel.AppLog.Errorf("Error occurred in NewAccessRequest: %s\nStack: \n%s", err, err.(stackTracer).StackTrace())
		oauth2Provider.WriteAccessError(rw, accessRequest, err)
		return
	}

	revel.AppLog.Infof("we have to create a token for: %+v", accessRequest)
	revel.AppLog.Infof("we have to create a token for: %+v", accessRequest.GetClient())
	revel.AppLog.Infof("we have to create a token for: %+v", accessRequest.GetSession())

	// Next we create a response for the access request. Again, we iterate through the TokenEndpointHandlers
	// and aggregate the result in response.
	response, err := oauth2Provider.NewAccessResponse(ctx, accessRequest)
	if err != nil {
		revel.AppLog.Errorf("Error occurred in NewAccessResponse: %s\nStack: \n%s", err, err.(stackTracer).StackTrace())
		oauth2Provider.WriteAccessError(rw, accessRequest, err)
		return
	}

	// if it is a token exchange, we save the id and the refresh token
	if accessRequest.GetGrantTypes().Has("authorization_code") {

		dbUser := dao.GetUser(accessRequest.GetSession().GetUsername())
		if dbUser == nil {
			revel.AppLog.Errorf("Logged In User %s nof found", dbUser.UserID)
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

// IntrospectionHandlerFunc
//  we check if Token is valid
func IntrospectionHandlerFunc(rw http.ResponseWriter, req *http.Request) {
	DoLogHTTPRequest(req, "Introspection-Handler")

	ctx := fosite.NewContext()
	mySessionData := newSession("")
	ir, err := oauth2Provider.NewIntrospectionRequest(ctx, req, mySessionData)
	if err != nil {
		revel.AppLog.Errorf("Error occurred in NewAuthorizeRequest: %+v\nStack: \n%s", err, err.(stackTracer).StackTrace())
		oauth2Provider.WriteIntrospectionError(rw, err)
		return
	}
	oauth2Provider.WriteIntrospectionResponse(rw, ir)
}

// RevokationHanlderFunc
func RevocationHandlerFunc(rw http.ResponseWriter, req *http.Request) {
	DoLogHTTPRequest(req, "Revocation-Handler")

	ctx := fosite.NewContext()
	err := oauth2Provider.NewRevocationRequest(ctx, req)
	oauth2Provider.WriteRevocationResponse(rw, err)
}

func CheckToken(token string) (active bool, username string) {
	// TODO: Maybe use API directly instead of using HTTP/Localhost

	if ok, _ := revel.Config.Bool("oauth.checktoken"); !ok {
		user, _ := revel.Config.String("user.admin")
		return true, user
	}

	// we build clientcredentials from servercredentials
	revel.AppLog.Debugf("We check Token for Validity: %s", token)

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
	revel.AppLog.Debugf("we get back from Introspection: %s", out)
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
	// refresh token expires in 10 years
	if tokenType == fosite.RefreshToken {
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
	result.Session = session //result.Session == nil -> dirty Fix
	data := dao.GetTokenBySignature(code)
	if data == nil {
		return nil, fosite.ErrNotFound
	}
	json.Unmarshal(*data, &result)
	return result, nil
}

func (c OAuthStorageAdapter) DeleteAuthorizeCodeSession(ctx context.Context, code string) (err error) {
	err = dao.DeleteToken(code)
	return err
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
	// TODO: check ! delete a single accesstoken
}

func (c OAuthStorageAdapter) CreateRefreshTokenSession(ctx context.Context, signature string, request fosite.Requester) (err error) {
	return storeToken(signature, fosite.RefreshToken, request)
}

func (c OAuthStorageAdapter) GetRefreshTokenSession(ctx context.Context, signature string, session fosite.Session) (request fosite.Requester, err error) {
	return c.GetAccessTokenSession(ctx, signature, session)
}

func (c OAuthStorageAdapter) DeleteRefreshTokenSession(ctx context.Context, signature string) (err error) {
	return c.DeleteAuthorizeCodeSession(ctx, signature)
	//TODO: check ! delete a refresh token and all access-tokens from the user
}

func (c OAuthStorageAdapter) RevokeRefreshToken(ctx context.Context, requestID string) error {
	return dao.DeleteTokenByTokenID(requestID, fosite.RefreshToken)
}

func (c OAuthStorageAdapter) RevokeAccessToken(ctx context.Context, requestID string) error {
	return dao.DeleteTokenByTokenID(requestID, fosite.AccessToken)
}

/*
TODO: not IMPLEMENTED
*/

func (c OAuthStorageAdapter) TokenRevocationStorage(ctx context.Context, requestID string) error {
	revel.AppLog.Infof("TokenRevocationStorage: %+v", ctx)
	return nil
}

func (c OAuthStorageAdapter) Authenticate(ctx context.Context, name string, secret string) error {
	revel.AppLog.Infof("Authenticate: %+v", ctx)
	return nil
}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
