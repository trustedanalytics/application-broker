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
		Operation("GetCatalog").
		Writes(catalog.CFCatalog{}))

	// service routes
	ws.Route(ws.PUT("/service_instances/{id}").
		To(sh.GetInstance).
		Param(ws.PathParameter("id", "service id").DataType("string")).
		Operation("GetInstance").
		Reads(catalog.CFServiceProvisioningResponse{}))

	rest.Add(ws)
}

func main() {
	s := BrokerService{}
	s.Initialize()
	u := fmt.Sprintf("%s:%s", s.Host, s.Port)
	log.Printf("server stats on: %s", u)
	log.Fatal(http.ListenAndServe(u, nil))
}
