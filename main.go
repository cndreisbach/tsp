package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"tailscale.com/client/tailscale"
	"tailscale.com/tailcfg"
	"tailscale.com/tsnet"
)

var (
	verbose  = flag.Bool("v", false, "be verbose")
	hostname = flag.String("h", "", "hostname for service")
	port     = flag.Int("p", 80, "port to proxy to")
)

var localClient *tailscale.LocalClient

func proxyHandler(p *httputil.ReverseProxy, a *url.URL) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var user *tailcfg.UserProfile

		res, err := localClient.WhoIs(r.Context(), r.RemoteAddr)
		if err == nil {
			user = res.UserProfile
		}

		if user != nil {
			r.Header.Set("X-WebAuth-User", user.LoginName)
			r.Header.Set("X-WebAuth-Name", user.DisplayName)
		}
		r.Host = a.Host

		w.Header().Set("X-Forwarded-By", "tsp")
		p.ServeHTTP(w, r)
	}
}

func tspRun() error {
	flag.Parse()

	if *hostname == "" {
		return errors.New("-h cannot be empty")
	}

	if flag.NArg() != 1 {
		return errors.New("you must specify a host to proxy to")
	}

	address, err := url.Parse(flag.Arg(0))
	if err != nil {
		return errors.New("you must specific a valid url as the host")
	}

	proxy := httputil.NewSingleHostReverseProxy(address)
	http.HandleFunc("/", proxyHandler(proxy, address))

	srv := &tsnet.Server{
		Hostname: *hostname,
		Logf:     func(format string, args ...any) {},
	}

	if *verbose {
		srv.Logf = log.Printf
	}

	if err := srv.Start(); err != nil {
		return err
	}

	localClient, _ = srv.LocalClient()

	l80, err := srv.Listen("tcp", ":"+strconv.Itoa(*port))
	if err != nil {
		return err
	}

	log.Printf("Serving %s as http://%s:%d/ ...", address, *hostname, *port)
	if err := http.Serve(l80, nil); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := tspRun(); err != nil {
		log.Fatal(err)
	}
}
