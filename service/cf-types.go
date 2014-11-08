package service

import "fmt"

// cfAppsResponse describes the Cloud Controller API result for a list of apps
type cfAppsResponse struct {
	Count     int             `json:"total_results"`
	Pages     int             `json:"total_pages"`
	Resources []cfAppResource `json:"resources"`
}

type cfAppResource struct {
	Meta   cfMeta `json:"metadata"`
	Entity cfApp  `json:"entity"`
}

type cfRoutesResponse struct {
	Count     int               `json:"total_results"`
	Pages     int               `json:"total_pages"`
	Resources []cfRouteResource `json:"resources"`
}

type cfRouteResource struct {
	Meta   cfMeta  `json:"metadata"`
	Entity cfRoute `json:"entity"`
}

type cfDomainResponse struct {
	Meta   cfMeta   `json:"metadata"`
	Entity cfDomain `json:"entity"`
}

type cfSpaceResource struct {
	Meta   cfMeta  `json:"metadata"`
	Entity cfSpace `json:"entity"`
}

type cfBindingResource struct {
	Meta   cfMeta            `json:"metadata"`
	Entity CFBindingResponse `json:"entity"`
}

type cfServiceResource struct {
	Meta   cfMeta    `json:"metadata"`
	Entity cfService `json:"entity"`
}

type cfMeta struct {
	GUID string `json:"guid"`
}

type cfService struct {
	GUID       string `json:"guid"`
	Name       string `json:"label"`
	Provider   string `json:"provider"`
	BrokerGUID string `json:"service_broker_guid"`
}

type cfApp struct {
	GUID      string `json:"guid"`
	Name      string `json:"name"`
	SpaceGUID string `json:"space_guid"`
	RoutesURL string `json:"routes_url"`
	// URI       string `json:"dashboard_url"`
}

type cfRoute struct {
	Host      string `json:"host"`
	DomainURL string `json:"domain_url"`
}

type cfDomain struct {
	Name string `json:"name"`
}

type cfSpace struct {
	GUID    string `json:"guid"`
	Name    string `json:"name"`
	OrgGUID string `json:"organization_guid"`
}

// CFServiceContext describes a CF Service Instance within the Cloud Controller
type CFServiceContext struct {
	InstanceID   string
	OrgName      string
	SpaceName    string
	InstanceName string
	AppName      string
}

// CFBindingResponse describes a CF Service Binding within the Cloud Controller
type CFBindingResponse struct {
	GUID                string `json:"guid"`
	AppGUID             string `json:"app_guid"`
	ServiceInstanceGUID string `json:"service_instance_guid"`
}

// NewCFServiceContext creates a new CFServiceContext including generated ServiceName
func NewCFServiceContext(instanceID string) (ctx *CFServiceContext) {
	ctx = &CFServiceContext{InstanceID: instanceID}
	ctx.AppName = fmt.Sprintf("%s-%s", Config.ServiceName, instanceID)
	return
}
