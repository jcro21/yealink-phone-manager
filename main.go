package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	// "strconv"
	"syscall"

	rice "github.com/GeertJohan/go.rice"
	"github.com/rs/cors"
	"github.com/sebest/xff"
	"github.com/urfave/negroni"
)

// Declare globals here
var (
	addr string
	// diskCriticalBytes int
)

var (
	canary = "blank"
)

var appCon *appContext

func init() {
	// Assign globals here
	addr = os.Getenv("ADDR")
	if addr == "" {
		addr = ":http"
	}

	// if i, err := strconv.Atoi(os.Getenv("DISK_CRITICAL_BYTES")); err != nil {
	// 	diskCriticalBytes = 104900000 // 100mb
	// } else {
	// 	diskCriticalBytes = i
	// }

	// these have to be here for rice to work
	_, _ = rice.FindBox("static")

	cfg := rice.Config{LocateOrder: []rice.LocateMethod{
		rice.LocateWorkingDirectory,
		rice.LocateFS,
		rice.LocateAppended,
	}}

	staticBox := cfg.MustFindBox("static")

	var appOptions []appOption

	appCon = newAppContext(
		staticBox,
		appOptions...,
	)
}

func main() {
	n := negroni.New()

	xffh, err := xff.Default()
	if err != nil {
		panic(err)
	}

	n.Use(xffh)
	n.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT"},
		AllowedHeaders:   []string{"accept", "accept-encoding", "authorization", "content-type"},
		AllowCredentials: true,
	}))

	n.UseHandler(appCon)

	fallbackHandler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		n.ServeHTTP(rw, r)
	})

	if os.Getenv("TLS") == "true" {
		fallbackHandler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			n.ServeHTTP(rw, r)
		})
	}

	s := http.Server{Addr: addr, Handler: fallbackHandler}

	ch := make(chan os.Signal, 1)

	signal.Notify(ch, syscall.SIGTERM)

	go func() {
		sig := <-ch

		fmt.Printf("\nshutting down in response to signal: %s", sig.String())

		if err := s.Shutdown(context.Background()); err != nil {
			panic(err)
		}
	}()

	fmt.Printf("Starting server on port %s", addr)

	if os.Getenv("TLS") == "true" {
		s2 := &http.Server{
			Handler:   n,
			Addr:      ":https",
			TLSConfig: &tls.Config{},
		}

		ch2 := make(chan os.Signal, 1)
		signal.Notify(ch2, syscall.SIGTERM)

		go func() {
			sig := <-ch2

			fmt.Printf("\nshutting down in response to signal: %s", sig.String())

			if err := s2.Shutdown(context.Background()); err != nil {
				panic(err)
			}
		}()

		fmt.Println("INFO Listening on 443 with TLS")
		go func() {
			if err := s2.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				panic(err)
			}
		}()
	}

	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}

	fmt.Println("successfully shut down")
}
