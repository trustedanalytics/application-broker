package main

import (
	"github.com/emicklei/go-restful"
	"github.com/intel-data/cf-catalog"
	"log"
	"net/http"
)

// CatalogProvider defines the required provider functionality
type CatalogProvider interface {
	Initialize() error
	GetCatalog() (*catalog.CFCatalog, error)
}

// CatalogHandler object
type CatalogHandler struct {
	Provider CatalogProvider
}

// Initialize configures the broker handler
func (h *CatalogHandler) Initialize() error {
	log.Println("initializing...")
	// TODO: Load the provider, is there a IOC pattern in go?
	c := &MockedCatalogProvider{}
	c.Initialize()
	h.Provider = c
	return nil
}

// GetCatalog returns a populated catalog for dynamically created services
func (h *CatalogHandler) GetCatalog(request *restful.Request, response *restful.Response) {
	log.Println("getting catalog...")
	c, err := h.Provider.GetCatalog()
	if err != nil {
		log.Printf("error on crating catalog: %v", err)
		response.WriteErrorString(
			http.StatusInternalServerError,
			"Error creating catalog")
	} else {
		response.WriteHeader(http.StatusOK)
		response.WriteEntity(c)
	}
}
