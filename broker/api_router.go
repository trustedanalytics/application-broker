package broker

import (
	"encoding/json"
	"log"

	"github.com/cloudfoundry-community/cf-go"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/oauth2"
)

func apiRouter(r martini.Router) {
	r.Get("/organizations", OrganizationsHandler)
	r.Get("/organizations/:org_id/spaces", OrganizationSpaceHandler)

}

func OrganizationsHandler(tokens oauth2.Tokens) []byte {
	config := cf.Config{ApiAddress: Config.ApiURL, Token: tokens.Access()}
	client := cf.NewClient(&config)
	orgs := client.ListOrganizations()
	orgsMarshal, err := json.Marshal(orgs)
	if err != nil {
		log.Printf("Error marshalling orgs %v", err)
	}
	return orgsMarshal
}

func OrganizationSpaceHandler(tokens oauth2.Tokens, params martini.Params) []byte {
	config := cf.Config{ApiAddress: Config.ApiURL, Token: tokens.Access()}
	client := cf.NewClient(&config)
	spaces := client.OrganizationSpaces(params["org_id"])
	spacesMarshal, err := json.Marshal(spaces)
	if err != nil {
		log.Printf("Error marshalling spaces %v", err)
	}
	return spacesMarshal
}
