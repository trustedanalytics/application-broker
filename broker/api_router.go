package broker

import (
	"encoding/json"
	"log"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/oauth2"
)

func apiRouter(r martini.Router) {
	r.Get("/organizations", OrgsHandler)
	r.Get("/organizations/:org_id/spaces", OrgSpaceHandler)
	r.Get("/apps/:app_id", AppHandler)
	r.Get("/apps", AppsHandler)
}

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

func AppHandler(tokens oauth2.Tokens, params martini.Params) []byte {
	config := cfclient.Config{ApiAddress: Config.ApiURL, Token: tokens.Access()}
	client := cfclient.NewClient(&config)
	app := client.AppByGuid(params["app_id"])
	appMarshal, err := json.Marshal(app)
	if err != nil {
		log.Printf("Error marshaling apps %v", err)
	}
	return appMarshal
}

func AppsHandler(tokens oauth2.Tokens) []byte {
	var returnApps []cfclient.App
	config := cfclient.Config{ApiAddress: Config.ApiURL, Token: tokens.Access()}
	client := cfclient.NewClient(&config)
	apps := client.ListApps()
	for _, app := range apps {
		if app.Environment["APP_LAUNCHER_NAME"] != "" {
			returnApps = append(returnApps, app)
		}
	}

	appsMarshal, err := json.Marshal(returnApps)
	if err != nil {
		log.Printf("Error marshaling apps %v", err)
	}
	return appsMarshal
}
