package main

import (
	"github.com/emicklei/go-restful"
	"log"
	"net/http"
)

func hasRequiredParams(req *restful.Request, res *restful.Response, args ...string) bool {
	for i, arg := range args {
		log.Printf("validating:%d - %v", i, arg)
		val := req.PathParameter(arg)
		if len(val) < 1 {
			log.Printf("nil %s", arg)
			res.WriteErrorString(
				http.StatusNotFound,
				"Required parameter not provided: "+arg)
			return false
		}
	}
	return true
}
