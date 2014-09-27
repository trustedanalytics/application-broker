package main

import (
	"flag"
	"fmt"
	rest "github.com/emicklei/go-restful"
	"github.com/intel-data/cf-catalog"
	"log"
	"net/http"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

// BrokerService object
type BrokerService struct {
	Host string
	Port string
}

// Initialize loads all broker service dependencies
func (s *BrokerService) Initialize() {

	// allows us to pass the host and port in
	flag.StringVar(&s.Host, "h", "0.0.0.0", "host")
	flag.StringVar(&s.Port, "p", "8888", "port")
	flag.Parse()

	ch := &CatalogHandler{}
	ch.Initialize()

	sh := &ServiceHandler{}
	sh.Initialize()

	ws := &rest.WebService{}
	ws.Path("/v2").
		Consumes(rest.MIME_JSON).
		Produces(rest.MIME_JSON)

	// catalog routes
	ws.Route(ws.GET("/catalog").
		To(ch.GetCatalog).
		Writes(catalog.CFCatalog{}))

	// service routes
	ws.Route(ws.PUT("/service_instances/{id}").
		To(sh.SetServiceInstance).
		Param(ws.PathParameter("id", "service id").DataType("string")).
		Reads(catalog.CFServiceProvisioningResponse{}))

	ws.Route(ws.PUT("/service_instances/{instId}/service_bindings/{bindId}").
		To(sh.SetServiceInstanceBinding).
		Param(ws.PathParameter("instId", "instance id").DataType("string")).
		Param(ws.PathParameter("bindId", "binding id").DataType("string")).
		Reads(catalog.CFServiceBindingRequest{}).
		Writes(catalog.CFServiceBindingResponse{}))

	ws.Route(ws.DELETE("/service_instances/{instId}/service_bindings/{bindId}").
		To(sh.DeleteServiceInstanceBinding).
		Param(ws.PathParameter("instId", "instance id").DataType("string")).
		Param(ws.PathParameter("bindId", "binding id").DataType("string")))

	ws.Route(ws.DELETE("/service_instances/{instId}").
		To(sh.DeleteServiceInstance).
		Param(ws.PathParameter("instId", "instance id").DataType("string")))

	rest.Add(ws)
}

func main() {
	s := BrokerService{}
	s.Initialize()
	u := fmt.Sprintf("%s:%s", s.Host, s.Port)
	log.Printf("server stats on: %s", u)
	log.Fatal(http.ListenAndServe(u, nil))
}
