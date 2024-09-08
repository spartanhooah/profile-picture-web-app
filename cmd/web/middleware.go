package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"reflect"
)

type contextKey string

const contextUserKey contextKey = "user_ip"

func (app *application) ipFromContext(ctx context.Context) string {
	return ctx.Value(contextUserKey).(string)
}

func (app *application) addIPToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		var ctx = context.Background()
		log.Println("Ctx is a", reflect.TypeOf(ctx))
		ip, err := getIP(req)

		if err != nil {
			log.Println("Got an error:", err)
			ip, _, _ = net.SplitHostPort(req.RemoteAddr)

			if len(ip) == 0 {
				ip = "unknown"
			}

			log.Println("First attempt, got an ip:", ip)

			ctx = context.WithValue(ctx, contextUserKey, ip)
		} else {
			log.Println("second attempt, got an ip:", ip)
			ctx = context.WithValue(ctx, contextUserKey, ip)
		}

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
