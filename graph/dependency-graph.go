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

package graph

import (
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/trustedanalytics/application-broker/cf-rest-api"
	"github.com/trustedanalytics/application-broker/types"
	"github.com/twmb/algoimpl/go/graph"
	"net/url"
	"strings"
)

type DependencyGraph struct {
	cf *api.CfAPI
}

func NewDependencyGraph() *DependencyGraph {
	toReturn := new(DependencyGraph)
	toReturn.cf = api.NewCfAPI()
	return toReturn
}

func (dg *DependencyGraph) addDependenciesToGraph(g *graph.Graph, parent graph.Node, sourceAppGUID string) error {
	log.Infof("addDependenciesToGraph for parent %v", *parent.Value)
	sourceAppSummary, err := dg.cf.GetAppSummary(sourceAppGUID)
	if err != nil {
		return err
	}
	for _, svc := range sourceAppSummary.Services {
		node := g.MakeNode()
		if dg.isNormalService(svc) {
			*node.Value = types.Component{
				GUID:         svc.GUID,
				Name:         svc.Name,
				Type:         types.ComponentService,
				DependencyOf: []string{(*parent.Value).(types.Component).GUID},
				Clone:        true,
			}
			g.MakeEdgeWeight(parent, node, 1)
		} else {
			*node.Value = types.Component{
				GUID:         svc.GUID,
				Name:         svc.Name,
				Type:         types.ComponentUPS,
				DependencyOf: []string{(*parent.Value).(types.Component).GUID},
				Clone:        true,
			}
			g.MakeEdgeWeight(parent, node, 1)
			// Retrieve UPS
			response, err := dg.cf.GetUserProvidedService(svc.GUID)
			if err != nil {
				return err
			}
			if val, ok := response.Entity.Credentials["url"]; ok {
				if urlStr, ok := val.(string); ok {
					appID, appName, err := dg.getAppIdAndNameFromSpaceByUrl(sourceAppSummary.SpaceGUID, urlStr)
					if err != nil {
						return err
					}
					if len(appID) > 0 {
						log.Infof("Application %v is bound using %v", appID, svc.Name)
						node2 := g.MakeNode()
						*node2.Value = types.Component{
							GUID:         appID,
							Name:         appName,
							Type:         types.ComponentApp,
							DependencyOf: []string{(*node.Value).(types.Component).GUID},
							Clone:        true,
						}
						g.MakeEdgeWeight(node, node2, 1)
						_ = dg.addDependenciesToGraph(g, node2, appID)
					}
				}
			}
		}
	}
	return nil
}

func (dg *DependencyGraph) graphHasCycles(g *graph.Graph) bool {
	components := g.StronglyConnectedComponents()
	for _, comp := range components {
		if len(comp) > 1 {
			return true
		}
	}
	return false
}

func (dg *DependencyGraph) mergeDuplicates(components []types.Component) []types.Component {
	m := make(map[string]int)
	var ret []types.Component
	for _, n := range components {
		j, ok := m[n.GUID]
		if !ok {
			ret = append(ret, n)
			m[n.GUID] = len(ret) - 1
			continue
		} else {
			log.Infof("Duplicated %v on position %v", n.Name, j)
			ret[j].DependencyOf = append(ret[j].DependencyOf, n.DependencyOf...)
			log.Infof("Merged dependencies %v", ret[j].DependencyOf)
		}
	}
	return ret
}

func (dg *DependencyGraph) isNormalService(svc types.CfAppSummaryService) bool {
	// Normal services require plan.
	// User provided services does not support Plans so this field is empty then.
	return len(svc.Plan.Service.Label) > 0
}

func (dg *DependencyGraph) getAppIdAndNameFromSpaceByUrl(spaceGUID, urlStr string) (string, string, error) {
	appURL, err := url.Parse(urlStr)
	if err != nil {
		log.Infof("[%v] is not a correct URL. Parsing failed.", urlStr)
		return "", "", err
	}
	log.Infof("URL Host %v", appURL.Host)
	routes, err := dg.cf.GetSpaceRoutesForHostname(spaceGUID, strings.Split(appURL.Host, ".")[0])
	if err != nil {
		return "", "", err
	}
	if routes.Count > 0 {
		log.Infof("%v route(s) retrieved for host %v", routes.Count, appURL.Host)
		routeGUID := routes.Resources[0].Meta.GUID
		apps, err := dg.cf.GetAppsFromRoute(routeGUID)
		if err != nil {
			return "", "", err
		}
		if apps.Count > 0 {
			app := apps.Resources[0]
			log.Infof("APP [%+v]", app)
			isSearched, err := dg.doesUrlMatchApplication(urlStr, app.Meta.GUID)
			if err != nil {
				return "", "", err
			}
			if isSearched {
				log.Infof("Found app match url in user provided service")
				return app.Meta.GUID, app.Entity.Name, nil
			} else {
				log.Infof("url of found app does not match url in user provided service")
			}
		} else {
			log.Infof("No apps bound to route: [%v]", routeGUID)
		}
	} else {
		log.Infof("No routes found for host: %v", appURL.Host)
	}
	return "", "", nil
}

func (dg *DependencyGraph) doesUrlMatchApplication(appUrlStr, appID string) (bool, error) {
	appURL, err := url.Parse(appUrlStr)
	if err != nil {
		return false, err
	}
	appSummary, err := dg.cf.GetAppSummary(appID)
	log.Infof("App summary retrieved is [%+v]", appSummary)
	if err != nil {
		return false, err
	}
	for i := range appSummary.Routes {
		route := appSummary.Routes[i]
		if appURL.Host == fmt.Sprintf("%v.%v", route.Host, route.Domain.Name) {
			return true, nil
		}
	}
	return false, nil
}
