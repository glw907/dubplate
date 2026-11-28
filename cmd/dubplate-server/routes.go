package main

import (
	"net/http"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/serve"
	"github.com/glw907/dubplate/internal/webapp"
)

// The subtree names this package mounts under the owned prefix, beside
// the boundary probe's [serve.ProbeRouteName] and the web app's names
// in webapp.
const (
	requestLinkRouteName = "request-link"
	consumeLinkRouteName = "consume-link"
	signOutRouteName     = "sign-out"

	// The shell hands the api mount the full path, so it cuts
	// <prefix>/api/v1/ itself and answers every other path under api
	// with a JSON 404. That is also why tus cannot live under api.
	apiRouteName = "api"
	tusRouteName = "tus"
)

type registrar interface {
	Register(name string, h http.Handler)
}

// mountRoutes is the registrar every route mounts through, never the
// shell directly, so the route inventory test sees each name.
var mountRoutes = func(s *serve.Shell) registrar { return s }

// registerAuthRoutes mounts the three auth handlers. They apply no
// method check of their own, so each mounts behind [methodOnly].
func registerAuthRoutes(routes registrar, h auth.Handlers) {
	routes.Register(requestLinkRouteName, methodOnly(http.MethodPost, h.RequestLink))
	routes.Register(consumeLinkRouteName, methodOnly(http.MethodPost, h.ConsumeLink))
	routes.Register(signOutRouteName, methodOnly(http.MethodPost, h.SignOut))
}

func registerWebRoutes(routes registrar, wr webapp.Routes) {
	routes.Register(webapp.LoginRouteName, wr.Login)
	routes.Register(webapp.AssetsRouteName, wr.Assets)
	routes.Register(webapp.FontsRouteName, wr.Fonts)
	routes.Register(webapp.FaviconSVGRouteName, wr.Icons)
	routes.Register(webapp.FaviconPNGRouteName, wr.Icons)
	routes.Register(webapp.MarkRouteName, wr.Icons)
	routes.Register(webapp.BrandingRouteName, wr.Branding)
	if wr.OperatorCSS != nil {
		routes.Register(webapp.OperatorCSSRouteName, wr.OperatorCSS)
	}
}

// registerAdminRoutes mounts wr's two admin documents behind
// [auth.RequireAdmin], outside the method and path checks, so a
// non-admin meets the gate's 404 whatever the method or path and only
// a verified admin meets a 405.
func registerAdminRoutes(routes registrar, wr webapp.Routes, admin auth.VerifyConfig) {
	routes.Register(webapp.MembersRouteName, auth.RequireAdmin(admin, wr.Members))
	routes.Register(webapp.ServerSettingsRouteName, auth.RequireAdmin(admin, wr.ServerSettings))
}

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
