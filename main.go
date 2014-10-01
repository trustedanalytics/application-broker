package main

import (
	"flag"
	"github.com/intel-data/app-launching-service-broker/broker"
	"github.com/intel-data/app-launching-service-broker/service"
	"log"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

func main() {
	log.Println("starting broker...")

	flag.Parse()

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
