package graph

import (
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/trustedanalytics/application-broker/cf-rest-api"
	"github.com/trustedanalytics/application-broker/misc"
	"github.com/trustedanalytics/application-broker/types"
	"github.com/twmb/algoimpl/go/graph"
)

type GraphAPI struct {
	c *api.CfAPI
}

func NewGraphAPI() *GraphAPI {
	toReturn := new(GraphAPI)
	toReturn.c = api.NewCfAPI()
	return toReturn
}

// Returns a list of services and apps which would be provisioned in normal run
func (gr *GraphAPI) DryRun(sourceAppGUID string) ([]types.Component, error) {
	sourceAppSummary, err := gr.c.GetAppSummary(sourceAppGUID)
	if err != nil {
		return nil, err
	}

	g := graph.New(graph.Directed)
	root := g.MakeNode()
	*root.Value = types.Component{
		GUID:         sourceAppGUID,
		Name:         sourceAppSummary.Name,
		Type:         types.ComponentApp,
		DependencyOf: []string{},
		Clone:        true,
	}

	dg := NewDependencyGraph()

	_ = dg.addDependenciesToGraph(g, root, sourceAppGUID)
	// Calculations
	sorted := g.TopologicalSort()
	log.Infof("Topological Order:\n")
	ret := make([]types.Component, len(sorted))
	for i, node := range sorted {
		text := ""
		for _, n := range g.Neighbors(node) {
			text += fmt.Sprint(*n.Value) + ","
		}
		log.Infof("%v [%v]", *node.Value, text)
		ret[len(sorted)-1-i] = (*node.Value).(types.Component)
	}
	ret = dg.mergeDuplicates(ret)

	if dg.graphHasCycles(g) {
		log.Errorf("Graph has cycles and stack cannot be copied")
		return nil, misc.InternalServerError{Context: "Graph has cycles and stack cannot be copied"}
	} else {
		log.Infof("Graph has no cycles")
	}
	return ret, nil
}
