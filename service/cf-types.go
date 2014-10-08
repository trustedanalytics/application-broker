package service

import ()

type CFAppsResponce struct {
	Count     int             `json:"total_results"`
	Pages     int             `json:"total_pages"`
	Resources []CFAppResource `json:"resources"`
}

type CFAppResource struct {
	Meta   CFMeta `json:"metadata"`
	Entity CFApp  `json:"entity"`
}

type CFSpaceResource struct {
	Meta   CFMeta  `json:"metadata"`
	Entity CFSpace `json:"entity"`
}

type CFMeta struct {
	GUID string `json:"guid"`
}

type CFApp struct {
	GUID      string `json:"guid"`
	Name      string `json:"name"`
	SpaceGUID string `json:"space_guid"`
}

type CFSpace struct {
	GUID    string `json:"guid"`
	Name    string `json:"name"`
	OrgGUID string `json:"organization_guid"`
}
