package broker

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/intel-data/types-cf"
	"log"
	"net/http"
)

// Broker to manage requests
type Broker struct {
	config         *Config
	catalogHandler *CatalogHandler
	serviceHandler *ServiceHandler
}

func New() *Broker {

	ch := &CatalogHandler{}
	ch.initialize()

	sh := &ServiceHandler{}
	sh.initialize()

	return &Broker{
		config:         BrokerConfig,
		catalogHandler: ch,
		serviceHandler: sh,
	}
}

func (s *Broker) Start() {

	ws := &restful.WebService{}
	ws.Path("/v2").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	// catalog routes
	ws.Route(ws.GET("/catalog").
		To(s.catalogHandler.getCatalog).
		Writes(cf.Catalog{}))

	// service routes
	ws.Route(ws.PUT("/service_instances/{serviceId}").
		To(s.serviceHandler.createService).
		Param(ws.PathParameter("serviceId", "service id").DataType("string")).
		Reads(cf.ServiceCreationRequest{}).
		Writes(cf.ServiceCreationResponce{}))

	ws.Route(ws.PUT("/service_instances/{serviceId}/service_bindings/{bindingId}").
		To(s.serviceHandler.createServiceBinding).
		Param(ws.PathParameter("serviceId", "service id").DataType("string")).
		Param(ws.PathParameter("bindingId", "binding id").DataType("string")).
		Reads(cf.ServiceBindingRequest{}).
		Writes(cf.ServiceBindingResponse{}))

	ws.Route(ws.DELETE("/service_instances/{serviceId}/service_bindings/{bindingId}").
		To(s.serviceHandler.deleteServiceBinding).
		Param(ws.PathParameter("serviceId", "service id").DataType("string")).
		Param(ws.PathParameter("bindingId", "binding id").DataType("string")))

	ws.Route(ws.DELETE("/service_instances/{serviceId}").
		To(s.serviceHandler.deleteService).
		Param(ws.PathParameter("serviceId", "service id").DataType("string")))

	restful.Add(ws)

	u := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	log.Printf("server: %s", u)
	log.Fatal(http.ListenAndServe(u, nil))

}

func handleServerError(response *restful.Response, err error) {
	log.Fatalf("internal server error: %s", err)
	response.WriteHeader(http.StatusInternalServerError)
	response.WriteErrorString(http.StatusInternalServerError,
		"Internal server error, please contact your platform operator")
}

func hasRequiredParams(req *restful.Request, res *restful.Response, args ...string) bool {
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
