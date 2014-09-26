package main

import (
	"github.com/emicklei/go-restful"
	"log"
	"net/http"
)

// CatalogHandler object
type CatalogHandler struct {
	Provider *CatalogProvider
}

// Initialize configures the broker handler
func (h *CatalogHandler) Initialize() error {
	log.Println("initializing...")
	// TODO: Load the provider, is there a IOC pattern in go?
	c := &CatalogProvider{}
	c.Initialize()
	h.Provider = c
	return nil
}

// GetCatalog returns a populated catalog for dynamically created services
func (h *CatalogHandler) GetCatalog(request *restful.Request, response *restful.Response) {
	log.Println("creating catalog...")
	c, err := h.Provider.NewCatalog()
	if err != nil {
		log.Printf("error on crating catalog: %v", err)
		response.WriteErrorString(
			http.StatusExpectationFailed,
			"Error creating catalog")
	} else {
		response.WriteEntity(c)
	}
}
