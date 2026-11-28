package main

import (
	"net/http"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/serve"
	"github.com/glw907/dubplate/internal/webapp"
)

// The subtree names serve mounts under the owned prefix, beside the
// boundary probe's [serve.ProbeRouteName].
const (
	requestLinkRouteName = "request-link"
	consumeLinkRouteName = "consume-link"
	signOutRouteName     = "sign-out"

	// The shell hands the api mount the full path, so it cuts
	// <prefix>/api/v1/ itself and answers every other path under api
	// with a JSON 404. That is also why tus cannot live under api.
	apiRouteName = "api"
	tusRouteName = "tus"

	// The two admin documents. The second is not "settings", a name
	// kept for the member preferences page.
	membersRouteName        = "members"
	serverSettingsRouteName = "server-settings"

	// The public web routes: the login documents, the bundle's hashed
	// assets, the two tier stylesheets that follow the bundle's own,
	// the embedded faces, the favicon in its two forms, and the
	// directory holding the mark's three detail levels.
	loginRouteName       = "login"
	assetsRouteName      = "_app"
	brandingRouteName    = "branding.css"
	operatorCSSRouteName = "operator.css"
	fontsRouteName       = "fonts"
	faviconSVGRouteName  = "favicon.svg"
	faviconPNGRouteName  = "favicon.png"
	markRouteName        = "mark"
)

// webRouteNames are the subtree names [webapp.New] builds its routes
// under, the same names [registerWebRoutes] mounts them at.
var webRouteNames = webapp.RouteNames{
	Login:          loginRouteName,
	Members:        membersRouteName,
	ServerSettings: serverSettingsRouteName,
	Assets:         assetsRouteName,
	Fonts:          fontsRouteName,
	FaviconSVG:     faviconSVGRouteName,
	FaviconPNG:     faviconPNGRouteName,
	Mark:           markRouteName,
	Branding:       brandingRouteName,
	OperatorCSS:    operatorCSSRouteName,
}

// A registrar mounts a route under a subtree name.
type registrar interface {
	Register(name string, h http.Handler)
}

// mountRoutes returns the registrar newServeShell mounts every route
// through. Every registration goes through it, never through the shell
// directly, so the route inventory test sees each name.
var mountRoutes = func(s *serve.Shell) registrar { return s }

// registerAuthRoutes mounts the three auth handlers. They apply no
// method check of their own, so each mounts behind [methodOnly].
func registerAuthRoutes(routes registrar, h auth.Handlers) {
	routes.Register(requestLinkRouteName, methodOnly(http.MethodPost, h.RequestLink))
	routes.Register(consumeLinkRouteName, methodOnly(http.MethodPost, h.ConsumeLink))
	routes.Register(signOutRouteName, methodOnly(http.MethodPost, h.SignOut))
}

// registerWebRoutes mounts wr's public routes.
func registerWebRoutes(routes registrar, wr webapp.Routes) {
	routes.Register(loginRouteName, wr.Login)
	routes.Register(assetsRouteName, wr.Assets)
	routes.Register(fontsRouteName, wr.Fonts)
	routes.Register(faviconSVGRouteName, wr.Icons)
	routes.Register(faviconPNGRouteName, wr.Icons)
	routes.Register(markRouteName, wr.Icons)
	routes.Register(brandingRouteName, wr.Branding)
	if wr.OperatorCSS != nil {
		routes.Register(operatorCSSRouteName, wr.OperatorCSS)
	}
}

// registerAdminRoutes mounts wr's two admin documents behind
// [auth.RequireAdmin], outside the method and path checks, so a
// non-admin meets the gate's 404 whatever the method or path and only
// a verified admin meets a 405.
func registerAdminRoutes(routes registrar, wr webapp.Routes, admin auth.VerifyConfig) {
	routes.Register(membersRouteName, auth.RequireAdmin(admin, wr.Members))
	routes.Register(serverSettingsRouteName, auth.RequireAdmin(admin, wr.ServerSettings))
}

// methodOnly wraps h, answering 405 with an Allow header naming method
// to any request whose method is not method.
func methodOnly(method string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.Header().Set("Allow", method)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		h.ServeHTTP(w, r)
	})
}
