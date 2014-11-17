package api

import (
	"encoding/json"
	"log"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/oauth2"
)

func OrgsHandler(tokens oauth2.Tokens) []byte {
	config := cfclient.Config{ApiAddress: Config.ApiURL, Token: tokens.Access()}
	client := cfclient.NewClient(&config)
	orgs := client.ListOrgs()
	orgsMarshal, err := json.Marshal(orgs)
	if err != nil {
		log.Printf("Error marshaling orgs %v", err)
	}
	return orgsMarshal
}

func OrgSpaceHandler(tokens oauth2.Tokens, params martini.Params) []byte {
	config := cfclient.Config{ApiAddress: Config.ApiURL, Token: tokens.Access()}
	client := cfclient.NewClient(&config)
	spaces := client.OrgSpaces(params["org_id"])
	spacesMarshal, err := json.Marshal(spaces)
	if err != nil {
		log.Printf("Error marshaling spaces %v", err)
	}
	return spacesMarshal
}
