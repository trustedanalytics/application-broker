package broker

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/go-martini/martini"
	auth2 "github.com/golang/oauth2"
	"github.com/intel-data/app-launching-service-broker/api"
	"github.com/martini-contrib/oauth2"
	"github.com/martini-contrib/sessions"
)

const (
	apiVersion = "v2"
	instanceID = "iid"
	bindingID  = "bid"
)

var (
	catalogURLPattern      = fmt.Sprintf("/%v/catalog", apiVersion)
	provisioningURLPattern = fmt.Sprintf("/%v/service_instances/:instance_id", apiVersion)
	bindingURLPattern      = fmt.Sprintf("/%v/service_instances/:instance_id/service_bindings/:binding_id", apiVersion)
)

type router struct {
	m *martini.ClassicMartini
}

func newRouter(h *handler) *router {

	m := martini.Classic()

	m.Use(sessions.Sessions("app_launcher", sessions.NewCookieStore([]byte("appsecretlauncher"))))
	m.Get(catalogURLPattern, reponseHandler(h.catalog))
	m.Put(provisioningURLPattern, reponseHandler(h.provision))
	m.Delete(provisioningURLPattern, reponseHandler(h.deprovision))
	m.Put(bindingURLPattern, reponseHandler(h.bind))
	m.Delete(bindingURLPattern, reponseHandler(h.unbind))
	if Config.UI {
		oauthOpts := &auth2.Options{
			ClientID:     Config.ClientID,
			ClientSecret: Config.ClientSecret,
			RedirectURL:  Config.RedirectURL,
			Scopes:       []string{""},
		}

		static := martini.Static("assets", martini.StaticOptions{Fallback: "/index.html", Exclude: "/v2"})

		cf := oauth2.NewOAuth2Provider(oauthOpts, Config.AuthURL, Config.TokenURL)

		m.Use(cf)
		m.Group("/ui", api.Router, oauth2.LoginRequired)

		m.NotFound(oauth2.LoginRequired, static, http.NotFound)
	}
	return &router{m}
}

// ServeHTTP logs all requests and dispatches to the appropriate handler
func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	// logging

	if Config.Debug {
		if dump, err := httputil.DumpRequest(req, true); err != nil {
			log.Printf("Cannot log incoming request: %v", err)
		} else {
			log.Print(string(dump))
		}
	}

	if Config.Debug {
		creds, err := parseCredentials(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		//TODO: should not print here, but until implemented
		log.Printf("Router: Authentication: [%v]", creds)
	}

	r.m.ServeHTTP(w, req)
}

type responseEntity struct {
	status int
	value  interface{}
}

type reponseHandler func(*http.Request, martini.Params) (int, string)

// ServeHTTP marshalls response as JSON, return the proper HTTP status code
func (fn reponseHandler) ServeHTTP(w http.ResponseWriter, req *http.Request, params martini.Params) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fn(req, params)
}

type credentials struct {
	usr string
	pas string
}

func parseCredentials(req *http.Request) (*credentials, error) {
	auths, _ := req.Header["Authorization"]
	if len(auths) != 1 {
		return nil, errors.New("Unauthorized access")
	}
	tokens := strings.Split(auths[0], " ")
	if len(tokens) != 2 || tokens[0] != "bearer" {
		return nil, errors.New("Unsupported authentication method")
	}
	raw, err := base64.StdEncoding.DecodeString(tokens[1])
	if err != nil {
		return nil, errors.New("Unable to decode 'Authorization' header")
	}
	creds := strings.Split(string(raw), ":")
	if len(creds) != 2 {
		return nil, errors.New("Missing credentials")
	}
	return &credentials{creds[0], creds[1]}, nil
}
