package service

import ()

type CFAppsResponce struct {
	Count     int             `json:"total_results"`
	Pages     int             `json:"total_pages"`
	Resources []CFAppResource `json:"resources"`
}

type CFServicesResponce struct {
	Resources []CFServiceResource `json:"resources"`
}

type CFAppResource struct {
	Meta   CFMeta `json:"metadata"`
	Entity CFApp  `json:"entity"`
}

type CFSpaceResource struct {
	Meta   CFMeta  `json:"metadata"`
	Entity CFSpace `json:"entity"`
}

type CFServiceResource struct {
	Meta   CFMeta    `json:"metadata"`
	Entity CFService `json:"entity"`
}

type CFMeta struct {
	GUID string `json:"guid"`
}

type CFService struct {
	GUID       string `json:"guid"`
	Name       string `json:"label"`
	Provider   string `json:"provider"`
	BrokerGUID string `json:"service_broker_guid"`
}

type CFApp struct {
	GUID      string `json:"guid"`
	Name      string `json:"name"`
	SpaceGUID string `json:"space_guid"`
	URI       string `json:"dashboard_url"`
}

type CFSpace struct {
	GUID    string `json:"guid"`
	Name    string `json:"name"`
	OrgGUID string `json:"organization_guid"`
}

type CFServiceContext struct {
	OrgName     string
	SpaceName   string
	ServiceName string
	ServiceURI  string
}
