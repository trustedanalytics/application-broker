package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/trustedanalytics/application-broker/misc"
	"github.com/trustedanalytics/application-broker/misc/http-utils"
	"github.com/trustedanalytics/application-broker/types"
	"net/http"
)

func (c *CfAPI) registerBroker(brokerName string, brokerURL string, username string, password string) error {
	address := fmt.Sprintf("%v/v2/service_brokers", c.BaseAddress)

	req := types.CfServiceBroker{Name: brokerName, URL: brokerURL, Username: username, Password: password}
	serialized, _ := json.Marshal(req)
	log.Infof("Registering broker: %v %v", address, serialized)

	request, err := http.NewRequest(httputils.MethodPost, address, bytes.NewReader(serialized))
	if err != nil {
		msg := fmt.Sprintf("Failed to prepare request for: %v %v", httputils.MethodPost, address)
		log.Error(msg)
		return misc.InternalServerError{Context: msg}
	}
	response, err := c.Do(request)

	if err != nil {
		msg := fmt.Sprintf("Failed to register service broker: %v", err.Error())
		log.Error(msg)
		return misc.InternalServerError{Context: msg}
	}

	if response.StatusCode != http.StatusCreated {
		msg := fmt.Sprintf("Failed to register service broker: Status code %d, Error %v", response.StatusCode,
			misc.ReaderToString(response.Body))
		log.Error(msg)
		return misc.InternalServerError{Context: msg}
	}

	return nil
}

func (c *CfAPI) updateBroker(brokerGUID string, brokerURL string, username string, password string) error {
	address := fmt.Sprintf("%v/v2/service_brokers/%v", c.BaseAddress, brokerGUID)

	req := types.CfServiceBroker{URL: brokerURL, Username: username, Password: password}
	serialized, _ := json.Marshal(req)

	log.Infof("Updating: %v %v", address, brokerURL)

	request, err := http.NewRequest(httputils.MethodPut, address, bytes.NewReader(serialized))
	if err != nil {
		msg := fmt.Sprintf("Failed to prepare request for: %v %v", httputils.MethodPost, address)
		log.Error(msg)
		return misc.InternalServerError{Context: msg}
	}
	response, err := c.Do(request)

	if err != nil {
		msg := fmt.Sprintf("Failed to update service broker: %v", err.Error())
		log.Error(msg)
		return misc.InternalServerError{Context: msg}
	}

	if response.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Failed to update service broker: Status code %d, Error %v", response.StatusCode,
			misc.ReaderToString(response.Body))
		log.Error(msg)
		return misc.InternalServerError{Context: msg}
	}

	return nil
}

func (c *CfAPI) getBrokers(brokerName string) (*types.CfServiceBrokerResources, error) {
	address := fmt.Sprintf("%v/v2/service_brokers?q=name:%v", c.BaseAddress, brokerName)
	response, err := c.Get(address)
	if err != nil {
		msg := fmt.Sprintf("Failed to get available service brokers: %v", err.Error())
		log.Error(msg)
		return nil, misc.InternalServerError{Context: msg}
	}

	if response.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Failed to get available service brokers: Status code %d, Error %v", response.StatusCode,
			misc.ReaderToString(response.Body))
		log.Error(msg)
		return nil, misc.InternalServerError{Context: msg}
	}

	brokers := new(types.CfServiceBrokerResources)
	if err := json.NewDecoder(response.Body).Decode(brokers); err != nil {
		msg := fmt.Sprintf("Failed to parse broker list response: %v", err.Error())
		log.Error(msg)
		return nil, misc.InternalServerError{Context: msg}
	}
	return brokers, nil
}
