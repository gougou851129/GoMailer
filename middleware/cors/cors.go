package cors

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"GoMailer/common/key"
	"GoMailer/common/utils"
	"GoMailer/handler/endpoint"
	"GoMailer/handler/userapp"
	"GoMailer/log"
)

var (
	freeAPI = map[string]struct{}{
		"/api/shortcut":  {},
		"/api/mail/list": {},
	}
)

func CORS(r *mux.Router) func(http.Handler) http.Handler {
	// required so we don't get a code 405
	r.Methods(http.MethodOptions).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := freeAPI[r.URL.Path]; ok {
				writeAllowOrigin(w, "*")
			} else {
				ak := key.EPKeyFromRequest(r)
				ep, err := endpoint.FindByKey(ak)
				if err != nil {
					log.Error("got error when find host for CORS origin: end point not exist")
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				app, err := userapp.FindById(ep.UserAppId)
				if err != nil {
					log.Error("got error when find host for CORS origin: user app not exist")
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				origin := r.Header.Get("Origin")
				if !isAllowed(app.Host, origin) {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				writeAllowOrigin(w, origin)
			}

			if r.Method == http.MethodOptions {
				// we only need headers for OPTIONS request, no need to go down the handler chain
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isAllowed(hosts string, origin string) bool {
	if utils.IsBlankStr(hosts) || utils.IsBlankStr(origin) {
		return false
	}

	hosts = strings.TrimSpace(hosts)
	hosts = strings.ReplaceAll(hosts, " ", "")
	parts := strings.Split(hosts, ",")
	for _, a := range parts {
		if a == origin {
			return true
		}
	}

	return false
}

func writeAllowOrigin(w http.ResponseWriter, allowOrigin string) {
	w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "POST,GET,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept-Encoding, User-Agent, Accept")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "86400")
}
