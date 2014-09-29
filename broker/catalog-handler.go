package broker

import (
	"github.com/emicklei/go-restful"
	"github.com/intel-data/generic-cf-service-broker/common"
	"github.com/intel-data/generic-cf-service-broker/service"
	"log"
	"net/http"
)

// CatalogHandler object
type CatalogHandler struct {
	provider common.CatalogProvider
}

func (h *CatalogHandler) initialize() error {
	log.Println("initializing...")
	// TODO: Load the provider, is there a IOC pattern in go?
	c := &service.MockedCatalogProvider{}
	c.Initialize()
	h.provider = c
	return nil
}

func (h *CatalogHandler) getCatalog(request *restful.Request, response *restful.Response) {
	log.Println("getting catalog...")
	c, err := h.provider.GetCatalog()
	if err != nil {
		handleServerError(response, err)
	} else {
		response.WriteHeader(http.StatusOK)
		response.WriteEntity(c)
	}
}
