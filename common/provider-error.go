package common

import (
	"fmt"
)

const (
	// ErrorServiceExists raised if instance already exists
	ErrorInstanceExists = 409
	// ErrorInstanceNotFound raised if instance not found
	ErrorInstanceNotFound = 410
	// ErrorException raised on server side error
	ErrorException = 500
)

func NewServiceProviderError(code int32, err error) *ServiceProviderError {
	return &ServiceProviderError{Code: code, Detail: err}
}

// ServiceProviderError describes service provider error
type ServiceProviderError struct {
	Code   int32
	Detail error
}

func (e *ServiceProviderError) String() string {
	return fmt.Sprintf("Error: %d (%s) - %v",
		e.Code, GetServiceProviderErrorCodeName[e.Code], e.Detail.Error())
}

// GetErrorCodeName resolves error code to its string value
var GetServiceProviderErrorCodeName = map[int32]string{
	409: "ErrorInstanceExists",
	410: "ErrorInstanceNotFound",
	500: "ErrorException",
}

// GetErrorCode resolves error name to its code
var GetServiceProviderErrorCode = map[string]int32{
	"ErrorInstanceExists":   409,
	"ErrorInstanceNotFound": 410,
	"ErrorException":        500,
}
