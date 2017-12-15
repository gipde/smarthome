package app

import (
	"encoding/json"
	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/pkg/errors"
	"github.com/revel/revel"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var (
	// Clients which will be allowed to connect to our Oauth2 Service
	clients map[string]*fosite.DefaultClient

	// die dürfen nur für kurze zeit verfügbar sein --> cleaner
	// ggf. über die Session abgefackelt -> es kommt ein Delete
	authorizeCodes = map[string]fosite.Requester{}
	// das könnte ggf. auch in die DB rein

	accessTokens           = map[string]fosite.Requester{}
	refreshTokens          = map[string]fosite.Requester{}
	accessTokenRequestIDs  = map[string]string{}
	refreshTokenRequestIDs = map[string]string{}
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

var (
	config = new(compose.Config)
	strg   = OAuthStorageAdapter{}

	strat = compose.CommonStrategy{
		CoreStrategy: compose.NewOAuth2HMACStrategy(config, []byte("some-super-cool-42-secret-that-nobody-knows")),
	}

	oauth2Provider = compose.Compose(
		config,
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
)

func init() {
	revel.OnAppStart(initOauth2)
}

func initOauth2() {
	var endpoint string
	endpoint, ok := revel.Config.String("oauth.endpoints")
	if !ok {
		os.Exit(1)
	}
	file, err := ioutil.ReadFile(revel.BasePath + "/" + endpoint)
	if err != nil {
		revel.AppLog.Errorf("File error: %v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(file, &clients)
}

func newSession(user string) *fosite.DefaultSession {
	return &fosite.DefaultSession{
		ExpiresAt: map[fosite.TokenType]time.Time{
			fosite.AccessToken:   time.Now().Add(time.Hour),
			fosite.AuthorizeCode: time.Now().Add(time.Hour),
		},
		Username: user,
	}
}

func AuthorizeHandlerFunc(rw http.ResponseWriter, req *http.Request) {
	ctx := fosite.NewContext()

	ar, err := oauth2Provider.NewAuthorizeRequest(ctx, req)
	if err != nil {
		revel.AppLog.Errorf("Error occurred in NewAuthorizeRequest: %s\nStack: \n%s", err, err.(stackTracer).StackTrace())
		oauth2Provider.WriteAuthorizeError(rw, ar, err)
		return
	}

	// Check a valid Revel-Session
	// Session is encrypted, so we can trust
	cook, _ := req.Cookie("REVEL_SESSION")
	session := revel.GetSessionFromCookie(revel.GoCookie(*cook))

	if user, ok := session["userid"]; ok {
		//TODO: hier sollten wir noch in die DB schreiben, dass es für den User eine Autorisierung für einen fremden Service gibt
		createAuthorizeResponse(ctx, ar, rw, user)
		return
	}

	// Append Parameters to Login-Page for further Redirect back to here
	pars := "client_id=" + req.Form.Get("client_id") + "&" +
		"redirect_uri=" + req.Form.Get("redirect_uri") + "&" +
		"response_type=" + req.Form.Get("response_type") + "&" +
		"scope=" + req.Form.Get("scope") + "&" +
		"state=" + req.Form.Get("state") + "&" +
		"none=" + req.Form.Get("none")

	// No Valid-Session -> Redirect to Resource-Server
	http.Redirect(rw, req, ContextRoot+"/?checkOauth2&"+pars, 302)
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

	// Awesome, now we redirect back to the client redirect uri and pass along an authorize code
	oauth2Provider.WriteAuthorizeResponse(rw, ar, response)
}

// Handler für alle Token Requests (authorize,revoke, ...)
func TokenHandlerFunc(rw http.ResponseWriter, req *http.Request) {
	// revel.AppLog.Infof("oauthRequest(Token): %+v", req)
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

	// TODO: hier die Daten in die DB reinschreiben
	// Da ist der Benutzername
	// accessRequest.GetSession().GetUsername()

	// Da ist der Token
	// response.GetAccessToken()

	// All done, send the response.
	oauth2Provider.WriteAccessResponse(rw, accessRequest, response)

	// The client now has a valid access token

}

// we check if token is valid
func IntrospectionEndpoint(rw http.ResponseWriter, req *http.Request) {
	ctx := fosite.NewContext()
	mySessionData := newSession("")
	ir, err := oauth2Provider.NewIntrospectionRequest(ctx, req, mySessionData)
	if err != nil {
		revel.AppLog.Errorf("Error occurred in NewAuthorizeRequest: %s\nStack: \n%s", err, err.(stackTracer).StackTrace())
		oauth2Provider.WriteIntrospectionError(rw, err)
		return
	}
	oauth2Provider.WriteIntrospectionResponse(rw, ir)
}

// Das muss in die DB rein

type OAuthStorageAdapter struct{}

func (c OAuthStorageAdapter) GetClient(ctx context.Context, id string) (fosite.Client, error) {
	// Jede Seite von der zugegriffen wird gilt als Client und braucht daher einen Eintrag
	revel.AppLog.Info("OAuth2 Request from client: ", id)
	cl, ok := clients[id]
	if !ok {
		return nil, fosite.ErrNotFound
	}
	return cl, nil
}

func (c OAuthStorageAdapter) CreateAuthorizeCodeSession(ctx context.Context, code string, request fosite.Requester) (err error) {
	revel.AppLog.Infof("CreateAuthorizeCodeSession code: %+v, request: %+v", code, request)
	authorizeCodes[code] = request
	return nil
	// in die DB schreiben
}

func (c OAuthStorageAdapter) GetAuthorizeCodeSession(ctx context.Context, code string, session fosite.Session) (request fosite.Requester, err error) {
	rel, ok := authorizeCodes[code]
	if !ok {
		return nil, fosite.ErrNotFound
	}
	return rel, nil

	// TODO: aus der DB laden
}

func (c OAuthStorageAdapter) DeleteAuthorizeCodeSession(ctx context.Context, code string) (err error) {
	delete(authorizeCodes, code)
	return nil

	// TODO: aus der DB löschen
}

func (c OAuthStorageAdapter) CreateAccessTokenSession(ctx context.Context, signature string, request fosite.Requester) (err error) {
	revel.AppLog.Infof("CreateAccessTokenSession code: %+v, request: %+v", signature, request)

	// dao.SaveToken() in DB, aber nur die signatur, und ggf. session.claim.audience

	accessTokens[signature] = request
	accessTokenRequestIDs[request.GetID()] = signature
	return nil
}

/*
 */
func (c OAuthStorageAdapter) GetAccessTokenSession(ctx context.Context, signature string, session fosite.Session) (request fosite.Requester, err error) {
	revel.AppLog.Info("GetAccessTokenSession: %+v", ctx)
	// das wird wohl über den Introspect benötigt
	// aus der DB laden
	return nil, nil
}

func (c OAuthStorageAdapter) DeleteAccessTokenSession(ctx context.Context, signature string) (err error) {
	revel.AppLog.Info("DeleteAccessTokenSession: %+v", ctx)
	return nil
}

func (c OAuthStorageAdapter) CreateRefreshTokenSession(ctx context.Context, signature string, request fosite.Requester) (err error) {
	revel.AppLog.Info("CreateRefreshTOkenSession: %+v", ctx)
	return nil
}

func (c OAuthStorageAdapter) GetRefreshTokenSession(ctx context.Context, signature string, session fosite.Session) (request fosite.Requester, err error) {
	revel.AppLog.Info("GetRefreshTOkenSession: %+v", ctx)
	return nil, nil
}

func (c OAuthStorageAdapter) DeleteRefreshTokenSession(ctx context.Context, signature string) (err error) {
	revel.AppLog.Info("DeleteRefreshTOkenSession: %+v", ctx)
	return nil
}

func (c OAuthStorageAdapter) TokenRevocationStorage(ctx context.Context, requestID string) error {
	revel.AppLog.Info("TokenRevocateionStorage: %+v", ctx)
	return nil
}

func (c OAuthStorageAdapter) RevokeRefreshToken(ctx context.Context, requestID string) error {
	revel.AppLog.Info("RevokeRefreshToken: %+v", ctx)
	return nil
}

func (c OAuthStorageAdapter) RevokeAccessToken(ctx context.Context, requestID string) error {
	revel.AppLog.Info("RevokeAccessToken: %+v", ctx)
	return nil
}

func (c OAuthStorageAdapter) Authenticate(ctx context.Context, name string, secret string) error {
	revel.AppLog.Info("Authenticate: %+v", ctx)
	return nil
}
