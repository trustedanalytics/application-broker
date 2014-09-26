package main

import (
	"flag"
	"fmt"
	rest "github.com/emicklei/go-restful"
	catalog "github.com/intel-data/cf-catalog"
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

	flag.StringVar(&s.Host, "h", "0.0.0.0", "host")
	flag.StringVar(&s.Port, "p", "8888", "port")
	flag.Parse()

	h := &BrokerHandlers{}
	ws := &rest.WebService{}
	ws.Path("/v2").
		Consumes(rest.MIME_JSON).
		Produces(rest.MIME_JSON)

	ws.Route(ws.GET("/catalog").To(h.GetCatalog).
		Doc("get a catalog").
		Operation("GetCatalog").
		Writes(catalog.CFCatalog{}))

	rest.Add(ws)
}

func main() {
	s := BrokerService{}
	s.Initialize()
	u := fmt.Sprintf("%s:%s", s.Host, s.Port)
	log.Printf("server stats on: %s", u)
	log.Fatal(http.ListenAndServe(u, nil))
}
