package web

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
)

type contextKey string

const contextUserKey contextKey = "user_ip"

func (app *Application) ipFromContext(ctx context.Context) string {
	return ctx.Value(contextUserKey).(string)
}

func (app *Application) AddIPToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		ip, err := getIP(req)

		if err != nil {
			log.Println("Got an error:", err)
			ip, _, _ = net.SplitHostPort(req.RemoteAddr)

			if len(ip) == 0 {
				ip = "unknown"
			}
		}

		ctx := context.WithValue(req.Context(), contextUserKey, ip)
		next.ServeHTTP(resp, req.WithContext(ctx))
	})
}

func getIP(req *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)

	if err != nil {
		return "unknown", err
	}

	userIP := net.ParseIP(ip)

	if userIP == nil {
		return "", fmt.Errorf("%q is not a valid IP", ip)
	}

	// check for a proxy
	forward := req.Header.Get("X-Forwarded-For")

	if len(forward) > 0 {
		ip = forward
	}

	if len(ip) == 0 {
		ip = "forward"
	}

	return ip, nil
}

func (app *Application) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if !app.Session.Exists(req.Context(), "user") {
			app.Session.Put(req.Context(), "error", "Log in first!")

			http.Redirect(resp, req, "/", http.StatusTemporaryRedirect)
			return
		}

		next.ServeHTTP(resp, req)
	})
}
