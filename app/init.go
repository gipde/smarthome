package app

import (
	"github.com/revel/revel"
	"net/http"
	"schneidernet/smarthome/app/dao"
	"time"
)

var (
	// AppVersion revel app version (ldflags)
	AppVersion string

	// BuildTime revel app build-time (ldflags)
	BuildTime string
)

func init() {
	revel.AppLog.Info("Initializing...")
	// Filters is the default set of global filters.
	revel.Filters = []revel.Filter{
		revel.PanicFilter,             // Recover from panics and display an error page instead.
		revel.RouterFilter,            // Use the routing table to select the right Action
		revel.FilterConfiguringFilter, // A hook for adding or removing per-Action filters.
		revel.ParamsFilter,            // Parse parameters into Controller.Params.
		revel.SessionFilter,           // Restore and write the session cookie.
		revel.FlashFilter,             // Restore and write the flash cookie.
		revel.ValidationFilter,        // Restore kept validation errors and save new ones from cookie.
		revel.I18nFilter,              // Resolve the requested language
		HeaderFilter,                  // Add some security based headers
		revel.InterceptorFilter,       // Run interceptors around the action.
		revel.CompressFilter,          // Compress the result.
		revel.ActionInvoker,           // Invoke the action.
	}

	// Register startup functions with OnAppStart
	// revel.DevMode and revel.RunMode only work inside of OnAppStart. See Example Startup Script
	// ( order dependent )
	revel.OnAppStart(installHandlers)
	revel.OnAppStart(ExampleStartupScript)
	revel.OnAppStart(dao.InitDB)

	// revel.OnAppStart(FillCache)
	revel.AppLog.Info("Initializing Ready")
	//revel.AppLog.Info("Current Engin: " + revel.CurrentEngine.Name())

}

// HeaderFilter adds common security headers
// There is a full implementation of a CSRF filter in
// https://github.com/revel/modules/tree/master/csrf
var HeaderFilter = func(c *revel.Controller, fc []revel.Filter) {
	c.Response.Out.Header().Add("X-Frame-Options", "SAMEORIGIN")
	c.Response.Out.Header().Add("X-XSS-Protection", "1; mode=block")
	c.Response.Out.Header().Add("X-Content-Type-Options", "nosniff")

	fc[0](c, fc[1:]) // Execute the next filter stage.
}

func ExampleStartupScript() {
	go markForever()
}

func markForever() {
	for {
		revel.AppLog.Info("--MARK--")
		time.Sleep(30 * time.Second)
	}
}

// func myHandler(w http.ResponseWriter, req *http.Request) {
// 	fmt.Fprintf(w, "Welcome to my Handler!")
// }

func installHandlers() {
	revel.AddInitEventHandler(func(event int, _ interface{}) (r int) {
		if event == revel.ENGINE_STARTED {

			srvHandler := &revel.CurrentEngine.(*revel.GoHttpServer).Server.Handler
			revelHandler := *srvHandler

			serveMux := http.NewServeMux()
			serveMux.Handle("/", revelHandler) // the old handler
			serveMux.Handle("/oauth2/auth",
				http.HandlerFunc(AuthorizeHandlerFunc)) // the oauth2
			serveMux.Handle("/oauth2/token",
				http.HandlerFunc(TokenHandlerFunc)) // the oauth2

			*srvHandler = serveMux
		}
		return
	})
}
