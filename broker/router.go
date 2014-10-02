package broker

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
)

const (
	apiVersion = "v2"
	instanceID = "iid"
	bindingID  = "bid"
)

var (
	catalogURLPattern      = fmt.Sprintf("/%v/catalog", apiVersion)
	provisioningURLPattern = fmt.Sprintf("/%v/service_instances/{%v}", apiVersion, instanceID)
	bindingURLPattern      = fmt.Sprintf("/%v/service_instances/{%v}/service_bindings/{%v}", apiVersion, instanceID, bindingID)
)

type router struct {
	mux *mux.Router
}

func newRouter(h *handler) *router {
	mux := mux.NewRouter()
	mux.Handle(catalogURLPattern, reponseHandler(h.catalog)).Methods("GET")
	mux.Handle(provisioningURLPattern, reponseHandler(h.provision)).Methods("PUT")
	mux.Handle(provisioningURLPattern, reponseHandler(h.deprovision)).Methods("DELETE")
	mux.Handle(bindingURLPattern, reponseHandler(h.bind)).Methods("PUT")
	mux.Handle(bindingURLPattern, reponseHandler(h.unbind)).Methods("DELETE")
	return &router{mux}
}

// ServeHTTP logs all requests and dispatches to the appropriate handler
func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	// logging
	if dump, err := httputil.DumpRequest(req, true); err != nil {
		log.Printf("Cannot log incoming request: %v", err)
	} else {
		log.Print(string(dump))
	}

	major, minor, err := parseApiVersion(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//TODO: Verify compatibility
	log.Printf("Router: Version check: [%v.%v]", major, minor)

	creds, err := parseCredentials(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	//TODO: should not print here, but until implemented
	log.Printf("Router: Authentication: [%v]", creds)

	r.mux.ServeHTTP(w, req)
}

type responseEntity struct {
	status int
	value  interface{}
}

type reponseHandler func(*http.Request) responseEntity

// ServeHTTP marshalls response as JSON, return the proper HTTP status code
func (fn reponseHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	re := fn(req)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(re.status)
	if err := json.NewEncoder(w).Encode(re.value); err != nil {
		log.Printf("Error on marshalling response: %v", err)
	}
}

func parseApiVersion(req *http.Request) (int, int, error) {
	versions, _ := req.Header["X-Broker-Api-Version"]
	if len(versions) != 1 {
		return 0, 0, errors.New("Missing Broker API version")
	}
	tokens := strings.Split(versions[0], ".")
	if len(tokens) != 2 {
		return 0, 0, errors.New("Invalid Broker API version")
	}
	major, err1 := strconv.Atoi(tokens[0])
	minor, err2 := strconv.Atoi(tokens[1])
	if err1 != nil || err2 != nil {
		return 0, 0, errors.New("Invalid Broker API version")
	}
	return major, minor, nil
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
	if len(tokens) != 2 || tokens[0] != "Basic" {
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
