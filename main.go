package main

import (
	"log"

	"github.com/intel-data/app-launching-service-broker/broker"
	"github.com/intel-data/app-launching-service-broker/service"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

func main() {

	s, err := service.New()
	if err != nil {
		log.Panicf("failed to initialize service: %v", err)
	}

	b, err := broker.New(s)
	if err != nil {
		log.Panicf("failed to initialize broker: %v", err)
	}

	b.Start()
}
