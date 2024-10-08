package web

import (
	"github.com/alexedwards/scs/v2"
	"net/http"
	"time"
)

func GetSession() *scs.SessionManager {
	session := scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true

	return session
}
