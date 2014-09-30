package broker

import (
	"fmt"
	sr "github.com/emicklei/go-restful"
	"github.com/intel-data/types-cf"
	"log"
	"net/http"
)

// Broker to manage requests
type Broker struct {
	config  *Config
	handler *Handler
}

func New(p cf.ServiceProvider) (*Broker, error) {
	h, err := NewHandler(p)
	if err != nil {
		log.Fatalf("error while creating handler: %v", err)
		return nil, err
	}

	return &Broker{
		config:  BrokerConfig,
		handler: h,
	}, nil
}

func (s *Broker) Start() {

	ws := &sr.WebService{}
	ws.Path("/v2").
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

	u := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	log.Printf("server: %s", u)
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
