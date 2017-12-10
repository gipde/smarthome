package tests

import (
	"github.com/revel/revel"
	"github.com/revel/revel/testing"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"strings"
)

type AppTest struct {
	testing.TestSuite
}

func (t *AppTest) Before() {
	println("Set up")
}

func (t *AppTest) TestThatIndexPageWorks() {
	t.Get("/")
	t.AssertOk()
	t.AssertContentType("text/html; charset=utf-8")
}

func (t *AppTest) TestLogin() {
	t.PostCustom("http://localhost:9000/main/login",
		"application/x-www-form-urlencoded", strings.NewReader("username=unknown&password=secret")).Send()

	t.AssertContains("User unbekannt")
}
func (t *AppTest) TestLogin2() {
	t.PostCustom("http://localhost:9000/main/login",
		"application/x-www-form-urlencoded", strings.NewReader("username=admin&password=admin")).Send()

	t.AssertNotContains("login-submit")
	t.AssertNotContains("Password false")
	t.AssertNotContains("Not Found")
	t.AssertContains("admin")
}

func (t *AppTest) TestOAuth2() {

	/*
		Ablauf
		1. Die Clientwebap ruft den Resource-Server  mit clientid  usw. auf
			(LOGIN Seite -> AuthorizeHandlerFunc!!!)
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

	// muss mit der Config am Server übereinstimmen
	conf := &oauth2.Config{
		ClientID:     "my-client",
		ClientSecret: "foobar",
		RedirectURL:  "http://localhost:3846/callback",
		Scopes: []string{
			"devices",
		},
		Endpoint: oauth2.Endpoint{
			TokenURL: "http://localhost:9000/oauth2/token",
			AuthURL:  "http://localhost:9000/oauth2/auth",
		},
	}

	authUrl := conf.AuthCodeURL("some-random-state-foobar", oauth2.AccessTypeOffline)
	revel.AppLog.Infof("authUrl: %+v", authUrl)

	// 1. Auth Request gegen den AUTH-Server
	t.GetCustom(authUrl).Send()
	// Kann auch als Post formuliert werden

	// Es wird redirected auf den Ressource-Server eingeleitet (/main/oauth2)
	// so dass sich der Anwender entweder einloggen kann, oder die Session benutzen kann

	// In dem Fall ist noch keine Session vorhanden, also wird auf das Login-Panel verzweigt
	t.AssertOk()
	t.AssertContains("login-submit")

	// 2. Dann wird ein Credential eingegeben und erneut gegen oauth2/auth geschickt
	t.PostCustom("http://localhost:9000/main/login", "application/x-www-form-urlencoded",
		strings.NewReader("client_id=my-client&redirect_uri=http%3A%2F%2Flocalhost%3A3846%2Fcallback&response_type=code&scope=devices&state=some-random-state-foobar&nonce=some-random-nonce&username=admin&password=admin")).Send()

	// Ausschalten, dass Redirects nicht mehr weiterverfolgt werden
	t.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	// 3. Nun haben wir eine gültige Session und wir wollen erneut einen Auth-Thoken erfragen (wie bei 1.)
	t.GetCustom(authUrl).Send()

	// 4. Der Authorize-Handler gibt nun einen Authorize-Code zurück und ruft die Callback Seite mit dem Auth-Code als Parameter auf
	t.AssertStatus(302)

	// Extraktion der Redirect-URL
	redirectURL, err := url.Parse(t.Response.Header.Get("Location"))
	if err != nil {
		t.Assert(false)
	}
	// Estraktion des Authorize-Codes
	cod := redirectURL.Query().Get("code")

	// 5. Wir tauschen den Auth-Token gegen einen Access-Token
	tok, err := conf.Exchange(oauth2.NoContext, cod)
	if err != nil {
		t.Assert(false)
	}
	revel.AppLog.Infof("token: %+v", tok)

	t.Assert(tok.AccessToken != "")
	t.Assert(tok.Valid())

}

func (t *AppTest) After() {
	println("Tear down")
}
