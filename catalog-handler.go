package main

import (
	"github.com/emicklei/go-restful"
	"github.com/intel-data/types-cf"
	"log"
	"net/http"
)

// CatalogProvider defines the required provider functionality
type CatalogProvider interface {
	getCatalog() (*cf.Catalog, error)
}

// CatalogHandler object
type CatalogHandler struct {
	provider CatalogProvider
}

func (h *CatalogHandler) initialize() error {
	log.Println("initializing...")
	// TODO: Load the provider, is there a IOC pattern in go?
	c := &MockedCatalogProvider{}
	c.initialize()
	h.provider = c
	return nil
}

func (h *CatalogHandler) getCatalog(request *restful.Request, response *restful.Response) {
	log.Println("getting catalog...")
	c, err := h.provider.getCatalog()
	if err != nil {
		handleServerError(response, err)
	} else {
		response.WriteHeader(http.StatusOK)
		response.WriteEntity(c)
	}
}
