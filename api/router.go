package api

import "github.com/go-martini/martini"

func Router(r martini.Router) {
	r.Get("/organizations", OrgsHandler)
	r.Get("/organizations/:org_id/spaces", OrgSpaceHandler)
	r.Get("/apps/:app_id", AppHandler)
	r.Get("/apps", AppsHandler)
}
