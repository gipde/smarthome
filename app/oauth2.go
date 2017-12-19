package app

import (
	"crypto/rand"
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

/*
TODO: Check Token Revokation
*/

var (
	// Clients which will be allowed to connect to our Oauth2 Service
	clients map[string]*fosite.DefaultClient
)

type OAuthStorageAdapter struct{}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

var (
	oauth2Provider  fosite.OAuth2Provider
	smartHomeClient clientcredentials.Config
)

func initOauth2() {
	var endpoint string
	endpoint, ok := revel.Config.String("oauth.clients")
	if !ok {
		os.Exit(1)
	}
	file, err := ioutil.ReadFile(revel.BasePath + "/" + endpoint)
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
		AccessTokenLifespan:   time.Minute * 30,
		AuthorizeCodeLifespan: time.Minute * 5,
		IDTokenLifespan:       time.Minute * 5,
	}
	strg := OAuthStorageAdapter{}

	oauthPass, _ := revel.Config.String("oauth.secret")
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
	revel.AppLog.Infof("new OAUTH2 Authorize Request from %s", req.RemoteAddr)
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

	// Now that the user is authorized, we set up a session. When validating / looking up tokens, we additionally get
	// the session. You can store anything you want in it.

	// The session will be persisted by the store and made available when e.g. validating tokens or handling token endpoint requests.
	// The default OAuth2 and OpenID Connect handlers require the session to implement a few methods. Apart from that, the
	// session struct can be anything you want it to be.

	// scope have to be devices
	scopes := ar.GetRequestedScopes()
	if len(scopes) != 1 || !scopes.Exact("devices") {
		http.Error(rw, "you're not allowed to do that", http.StatusForbidden)
		return
	}

	// Now we need to get a response. This is the place where the AuthorizeEndpointHandlers kick in and start processing the request.
	// NewAuthorizeResponse is capable of running multiple response type handlers which in turn enables this library
	// to support open id connect.
	response, err := oauth2Provider.NewAuthorizeResponse(ctx, ar, mySessionData)
	if err != nil {
		oauth2Provider.WriteAuthorizeError(rw, ar, err)
		return
	}

	// Save Auth App in DB
	dbUser := dao.GetUser(user)
	if dbUser == nil {
		revel.AppLog.Errorf("Logged In User %s nof found", user)
		return
	}
	if dbUser.Authorizations == nil {
		dbUser.Authorizations = []dao.AuthorizeEntry{}
	}
	dbUser.Authorizations = append(dbUser.Authorizations, dao.AuthorizeEntry{
		AppID: ar.GetClient().GetID(),
	})
	dao.SaveUser(dbUser)

	// Awesome, now we redirect back to the client redirect uri and pass along an authorize code
	oauth2Provider.WriteAuthorizeResponse(rw, ar, response)
}

// Handler für alle Token Requests (authorize,revoke, ...)
func TokenHandlerFunc(rw http.ResponseWriter, req *http.Request) {
	revel.AppLog.Infof("new OAuth2 Token Request from %s", req.RemoteAddr)
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

	// Next we create a response for the access request. Again, we iterate through the TokenEndpointHandlers
	// and aggregate the result in response.
	response, err := oauth2Provider.NewAccessResponse(ctx, accessRequest)
	if err != nil {
		revel.AppLog.Errorf("Error occurred in NewAccessResponse: %s\nStack: \n%s", err, err.(stackTracer).StackTrace())
		oauth2Provider.WriteAccessError(rw, accessRequest, err)
		return
	}

	revel.AppLog.Debugf("We created a new ACCESSTOKEN for grants %+v: %s", accessRequest.GetGrantTypes(), response.GetAccessToken())

	// All done, send the response.
	oauth2Provider.WriteAccessResponse(rw, accessRequest, response)

	// The client now has a valid access token

}

// IntrospectionHandlerFunc
//  we check if Token is valid
func IntrospectionHandlerFunc(rw http.ResponseWriter, req *http.Request) {
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

func CheckToken(token string) (active bool, username string) {

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
	json.Unmarshal(out, &introspection)

	return introspection.Active, introspection.Username
}

/*
Oauth2 Storage Implementation
*/

func (c OAuthStorageAdapter) GetClient(ctx context.Context, id string) (fosite.Client, error) {
	cl, ok := clients[id]
	if !ok {
		return nil, fosite.ErrNotFound
	}
	return cl, nil
}

func (c OAuthStorageAdapter) CreateAuthorizeCodeSession(ctx context.Context, code string, request fosite.Requester) (err error) {
	serialized, _ := json.Marshal(request)
	revel.AppLog.Debugf("Create Auth-Request: %+v", string(serialized))
	expiry := request.GetSession().GetExpiresAt("access_token")
	err = dao.SaveToken(code, expiry, &serialized)
	return err
}

func (c OAuthStorageAdapter) GetAuthorizeCodeSession(ctx context.Context, code string, session fosite.Session) (request fosite.Requester, err error) {
	var result = fosite.NewAuthorizeRequest()
	result.Session = session //result.Session == nil -> dirty Fix
	data := dao.GetToken(code)
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
	serialized, _ := json.Marshal(request)
	expiry := request.GetSession().GetExpiresAt("access_token")
	err = dao.SaveToken(signature, expiry, &serialized)
	return err
}

func (c OAuthStorageAdapter) GetAccessTokenSession(ctx context.Context, signature string, session fosite.Session) (request fosite.Requester, err error) {
	var result = fosite.NewAccessRequest(session)
	data := dao.GetToken(signature)
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
	return c.CreateAccessTokenSession(ctx, signature, request)
}

func (c OAuthStorageAdapter) GetRefreshTokenSession(ctx context.Context, signature string, session fosite.Session) (request fosite.Requester, err error) {
	return c.GetAccessTokenSession(ctx, signature, session)
}

func (c OAuthStorageAdapter) DeleteRefreshTokenSession(ctx context.Context, signature string) (err error) {
	return c.DeleteAuthorizeCodeSession(ctx, signature)
}

/*
TODO: not IMPLEMENTED
*/

func (c OAuthStorageAdapter) TokenRevocationStorage(ctx context.Context, requestID string) error {
	revel.AppLog.Infof("TokenRevocateionStorage: %+v", ctx)
	return nil
}

func (c OAuthStorageAdapter) RevokeRefreshToken(ctx context.Context, requestID string) error {
	revel.AppLog.Infof("RevokeRefreshToken: %+v", ctx)
	return nil
}

func (c OAuthStorageAdapter) RevokeAccessToken(ctx context.Context, requestID string) error {
	revel.AppLog.Infof("RevokeAccessToken: %+v", ctx)
	return nil
}

func (c OAuthStorageAdapter) Authenticate(ctx context.Context, name string, secret string) error {
	revel.AppLog.Infof("Authenticate: %+v", ctx)
	return nil
}
