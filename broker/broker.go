package broker

import (
	"encoding/base64"
	"errors"
	"fmt"
	sr "github.com/emicklei/go-restful"
	"github.com/intel-data/types-cf"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	ApiVersion = "/v2"
)

// Broker to manage requests
type Broker struct {
	config  *BrokerConfig
	handler *Handler
}

func New(p cf.ServiceProvider) (*Broker, error) {
	h, err := NewHandler(p)
	if err != nil {
		log.Fatalf("error while creating handler: %v", err)
		return nil, err
	}

	return &Broker{
		config:  Config,
		handler: h,
	}, nil
}

func (s *Broker) Start() {

	ws := &sr.WebService{}
	ws.Path(ApiVersion).
		Consumes(sr.MIME_JSON).
		Produces(sr.MIME_JSON)

	// catalog routes
	ws.Route(ws.GET("/catalog").
		To(s.handler.getCatalog).
		Writes(cf.Catalog{}))

	// service routes
	ws.Route(ws.PUT("/service_instances/{instanceId}").
		To(s.handler.createService).
		Param(ws.PathParameter("instanceId", "instance id").DataType("string")).
		Reads(cf.ServiceCreationRequest{}).
		Writes(cf.ServiceCreationResponce{}))

	ws.Route(ws.PUT("/service_instances/{instanceId}/service_bindings/{bindingId}").
		To(s.handler.createServiceBinding).
		Param(ws.PathParameter("instanceId", "instance id").DataType("string")).
		Param(ws.PathParameter("bindingId", "binding id").DataType("string")).
		Reads(cf.ServiceBindingRequest{}).
		Writes(cf.ServiceBindingResponse{}))

	ws.Route(ws.DELETE("/service_instances/{instanceId}/service_bindings/{bindingId}").
		To(s.handler.deleteServiceBinding).
		Param(ws.PathParameter("instanceId", "instance id").DataType("string")).
		Param(ws.PathParameter("bindingId", "binding id").DataType("string")))

	ws.Route(ws.DELETE("/service_instances/{instanceId}").
		To(s.handler.deleteService).
		Param(ws.PathParameter("instanceId", "instance id").DataType("string")))

	sr.Add(ws)

	u := fmt.Sprintf(":%d", s.config.CFEnv.Port)
	log.Printf("server starts on port %d", s.config.CFEnv.Port)
	log.Fatal(http.ListenAndServe(u, nil))

}

func handleSimpleServerError(response *sr.Response, err error) {
	if err == nil {
		return
	}
	handleServerError(response, cf.NewServiceProviderError(cf.ErrorServerException, err))
}

func handleServerError(response *sr.Response, err *cf.ServiceProviderError) {
	if err == nil {
		return
	}

	log.Fatalf("internal server error: %s", err.String())

	switch err.Code {
	case cf.ErrorInstanceExists:
		response.WriteHeader(http.StatusConflict)
	case cf.ErrorInstanceNotFound:
		response.WriteHeader(http.StatusNotFound)
	default:
		response.WriteHeader(http.StatusInternalServerError)
		response.WriteErrorString(http.StatusInternalServerError,
			"Internal server error, please contact your platform operator")
	}

}

func hasRequiredParams(req *sr.Request, res *sr.Response, args ...string) bool {
	for i, arg := range args {
		log.Printf("validating:%d - %v", i, arg)
		val := req.PathParameter(arg)
		if len(val) < 1 {
			log.Printf("nil %s", arg)
			res.WriteErrorString(http.StatusBadRequest,
				"Required parameter not provided: "+arg)
			return false
		}
	}
	return true
}

// Thanks to michaljemala for this
func getVersion(req *sr.Request) (int, int, error) {
	v := req.HeaderParameter("X-Broker-Api-Version")
	if len(v) < 1 {
		return 0, 0, errors.New("Missing Broker API version")
	}
	tokens := strings.Split(v, ".")
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

func extractCredentials(req *sr.Request) (string, string, error) {
	auths := req.HeaderParameter("Authorization")
	if len(auths) < 1 {
		return "", "", errors.New("Unauthorized access")
	}
	tokens := strings.Split(auths, " ")
	if len(tokens) != 2 || tokens[0] != "Basic" {
		return "", "", errors.New("Unsupported authentication method")
	}
	raw, err := base64.StdEncoding.DecodeString(tokens[1])
	if err != nil {
		return "", "", errors.New("Unable to decode 'Authorization' header")
	}
	credentials := strings.Split(string(raw), ":")
	if len(credentials) != 2 {
		return "", "", errors.New("Missing credentials")
	}
	return credentials[0], credentials[1], nil
}
