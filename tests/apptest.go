package tests

import (
	"github.com/revel/revel"
	"github.com/revel/revel/testing"
	"schneidernet/smarthome/app"
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
	t.PostCustom(app.PublicHost+app.ContextRoot+"/main/login",
		"application/x-www-form-urlencoded", strings.NewReader("username=unknown&password=secret")).Send()

	t.AssertContains("User unbekannt")
}
func (t *AppTest) TestLogin2() {
	t.PostCustom(app.PublicHost+app.ContextRoot+"/main/login",
		"application/x-www-form-urlencoded", strings.NewReader("username=admin&password=admin")).Send()

	t.AssertNotContains("login-submit")
	t.AssertNotContains("Password false")
	t.AssertNotContains("Not Found")
	t.AssertContains("admin")
}

func (t *AppTest) TestRedirectToIntentionalPage() {
	// first we want to enter Oauth page
	t.GetCustom("http://localhost:8180/Main/Oauth").Send()
	// but we are not logged in
	t.AssertContains("login-submit")
	t.AssertStatus(200)
	// we send credentials to Login
	t.PostCustom("http://localhost:8180/Main/Login", "application/x-www-form-urlencoded",
		strings.NewReader("username=admin&password=admin")).Send()

	// now we should be on the Oauth page
	t.AssertStatus(200)
	revel.AppLog.Infof("Response %+v", t.Response)
	t.AssertContains("SmartHome Permission Request")

}

func (t *AppTest) After() {
	println("Tear down")
}
