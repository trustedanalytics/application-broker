package service

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
	URI       string `json:"dashboard_url"`
}

type cfSpace struct {
	GUID    string `json:"guid"`
	Name    string `json:"name"`
	OrgGUID string `json:"organization_guid"`
}

// CFServiceContext describes a CF Service Instance within the Cloud Controller
type CFServiceContext struct {
	InstanceID  string
	OrgName     string
	SpaceName   string
	ServiceName string
	ServiceURI  string
}

// CFBindingResponse describes a CF Service Binding within the Cloud Controller
type CFBindingResponse struct {
	GUID                string `json:"guid"`
	AppGUID             string `json:"app_guid"`
	ServiceInstanceGUID string `json:"service_instance_guid"`
}
