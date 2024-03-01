package main

import (
	"net/http"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gorilla/mux"
	"github.com/timewasted/go-accept-headers"
)

type appOption func(a *appContext)

type appContext struct {
	sb *rice.Box
	m  *mux.Router
}

func newAppContext(staticBox *rice.Box, options ...appOption) *appContext {
	a := &appContext{
		sb: staticBox,
		m:  mux.NewRouter().UseEncodedPath(),
	}

	for _, fn := range options {
		fn(a)
	}

	a.m.Methods("GET").Path("/health").HandlerFunc(a.handleAPIHealth)
	a.m.Methods("GET").Path("/y000000000028.cfg").HandlerFunc(a.handleAPIPhoneSettingsFileGet)
	a.m.Methods("POST").Path("/api/v1/phone/settings").Handler(tollbooth.LimitFuncHandler(tollbooth.NewLimiter(1, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Second * 2}), a.handleAPIPhoneSettingsPut))
	a.m.Methods("GET").Path("/favicon.ico").Handler(http.RedirectHandler("static/favicon.ico", http.StatusFound))
	a.m.Methods("GET").Path("/").HandlerFunc(a.handleAPIPhoneSettingsFrontendGet)

	a.m.PathPrefix("/static").Handler(http.StripPrefix("/static", http.FileServer(staticBox.HTTPBox())))

	return a
}

func (a *appContext) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	acceptable := accept.Parse(r.Header.Get("accept"))
	if acceptable.Accepts("text/css") {
		ct, err := acceptable.Negotiate("text/html", "application/json", "text/css")
		if err != nil {
			panic(err)
		}
		if ct == "text/css" {
			rw.Header().Set("content-type", "text/css")
		}
	}

	a.m.ServeHTTP(rw, r)
}
