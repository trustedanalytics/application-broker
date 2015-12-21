/**
 * Copyright (c) 2015 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package types

// cfAppsResponse describes the Cloud Controller API result for a list of apps
type CfAppsResponse struct {
	Count     int             `json:"total_results"`
	Pages     int             `json:"total_pages"`
	Resources []CfAppResource `json:"resources"`
}

type CfAppResource struct {
	Meta   CfMeta `json:"metadata"`
	Entity CfApp  `json:"entity"`
}

type CfAppSummary struct {
	CfApp
	GUID     string                `json:"guid"`
	Routes   []CfAppSummaryRoute   `json:"routes"`
	Services []CfAppSummaryService `json:"services"`
}

type CfAppSummaryService struct {
	GUID string                  `json:"guid"`
	Name string                  `json:"name"`
	Plan CfAppSummaryServicePlan `json:"service_plan"`
}

type CfAppSummaryServicePlan struct {
	GUID    string `json:"guid"`
	Name    string `json:"name"`
	Service CfAppSummaryServicePlanService
}

type CfAppSummaryServicePlanService struct {
	GUID  string `json:"guid"`
	Label string `json:"label"`
}

type CfRoutesResponse struct {
	Count     int               `json:"total_results"`
	Pages     int               `json:"total_pages"`
	Resources []CfRouteResource `json:"resources"`
}

type CfRouteResource struct {
	Meta   CfMeta  `json:"metadata"`
	Entity CfRoute `json:"entity"`
}

type CfAppSummaryRoute struct {
	GUID   string   `json:"guid"`
	Host   string   `json:"host"`
	Domain CfDomain `json:"domain"`
}

type CfDomainResponse struct {
	Meta   CfMeta   `json:"metadata"`
	Entity CfDomain `json:"entity"`
}

type CfSpaceResource struct {
	Meta   CfMeta  `json:"metadata"`
	Entity CfSpace `json:"entity"`
}

type CfBindingResource struct {
	Meta   CfMeta    `json:"metadata"`
	Entity CfBinding `json:"entity"`
}

type CfServiceResource struct {
	Meta   CfMeta    `json:"metadata"`
	Entity CfService `json:"entity"`
}

type CfMeta struct {
	GUID string `json:"guid"`
	URL  string `json:"url"`
}

type CfJob struct {
	GUID   string `json:"guid"`
	Status string `json:"status"`
	Error  string `json:"error"`
}

type CfJobResponse struct {
	Meta   CfMeta `json:"metadata"`
	Entity CfJob  `json:"entity"`
}

type CfService struct {
	GUID       string `json:"guid"`
	Name       string `json:"label"`
	Provider   string `json:"provider"`
	BrokerGUID string `json:"service_broker_guid"`
}

type CfApp struct {
	Name          string                 `json:"name"`
	SpaceGUID     string                 `json:"space_guid"`
	RoutesURL     string                 `json:"routes_url"`
	State         string                 `json:"state"`
	BuildpackUrl  string                 `json:"buildpack"`
	Command       string                 `json:"command"`
	DiskQuota     int64                  `json:"disk_quota"`
	InstanceCount int                    `json:"instances"`
	Memory        int64                  `json:"memory"`
	Path          string                 `json:"path"`
	Envs          map[string]interface{} `json:"environment_json"`
}

type ServiceBindingResponse struct {
	Credentials map[string]string `json:"credentials"`
}

type CfAppInstance struct {
	State string  `json:"state"`
	Since float64 `json:"since"`
}

type CfCopyBitsRequest struct {
	SrcAppGUID string `json:"source_app_guid"`
}

type CfRoute struct {
	Host       string `json:"host"`
	DomainURL  string `json:"domain_url"`
	DomainGUID string `json:"domain_guid"`
}

type CfCreateRouteRequest struct {
	Host       string `json:"host"`
	DomainGUID string `json:"domain_guid"`
	SpaceGUID  string `json:"space_guid"`
}

type CfDomain struct {
	GUID string `json:"guid"`
	Name string `json:"name"`
}

type CfSpace struct {
	GUID    string `json:"guid"`
	Name    string `json:"name"`
	OrgGUID string `json:"organization_guid"`
}

// CfServiceContext describes a CF Service Instance within the Cloud Controller
type CfServiceContext struct {
	InstanceID   string
	OrgName      string
	SpaceName    string
	SpaceGUID    string
	InstanceName string
	AppName      string
}

// CFBindingResponse describes a CF Service Binding within the Cloud Controller
type CfBinding struct {
	GUID                string `json:"guid"`
	AppGUID             string `json:"app_guid"`
	ServiceInstanceGUID string `json:"service_instance_guid"`
}

type CfServiceInstanceCreateRequest struct {
	Name      string                 `json:"name"`
	SpaceGUID string                 `json:"space_guid"`
	PlanGUID  string                 `json:"service_plan_guid,omitempty"`
	Params    map[string]interface{} `json:"parameters,omitempty"`
	Tags      []string               `json:"tags,omitempty"`
}

type CfServiceInstanceCreateResponse struct {
	Meta CfMeta `json:"metadata"`
}

type CfServiceBindingCreateRequest struct {
	ServiceInstanceGUID string                 `json:"service_instance_guid"`
	AppGUID             string                 `json:"app_guid"`
	Params              map[string]interface{} `json:"parameters,omitempty"`
}

type CfServiceBindingCreateResponse struct {
	Meta CfMeta `json:"metadata"`
}

type CfBindingsResources struct {
	TotalResults int                 `json:"total_results"`
	Resources    []CfBindingResource `json:"resources"`
}

type CfVcapApplication struct {
	Name string   `json:"name"`
	Uris []string `json:"uris"`
}

type CfServiceBrokerResources struct {
	TotalResults int                       `json:"total_results"`
	Resources    []CfServiceBrokerResource `json:"resources"`
}

type CfServiceBrokerResource struct {
	Meta   CfMeta          `json:"metadata"`
	Entity CfServiceBroker `json:"entity"`
}

type CfServiceBroker struct {
	Name     string `json:"name,omitempty"`
	URL      string `json:"broker_url"`
	Username string `json:"auth_username"`
	Password string `json:"auth_password"`
}

const (
	AppStarted = "STARTED"
	AppStopped = "STOPPED"
)

func NewCfAppResource(summary CfAppSummary, newName string, spaceGUID string) *CfAppResource {
	summary.CfApp.Name = newName
	summary.CfApp.State = AppStopped
	summary.CfApp.SpaceGUID = spaceGUID
	return &CfAppResource{Meta: CfMeta{GUID: summary.GUID}, Entity: summary.CfApp}
}

func NewCfServiceInstanceRequest(name string, spaceGUID string, plan CfAppSummaryServicePlan) *CfServiceInstanceCreateRequest {
	return &CfServiceInstanceCreateRequest{
		Name:      name,
		PlanGUID:  plan.GUID,
		SpaceGUID: spaceGUID,
		Tags:      []string{plan.Service.Label}}
}

func NewCfServiceBindingRequest(appGUID string, svcInstanceGUID string) *CfServiceBindingCreateRequest {
	return &CfServiceBindingCreateRequest{
		AppGUID:             appGUID,
		ServiceInstanceGUID: svcInstanceGUID}
}
