package main

import (
	"github.com/intel-data/generic-cf-service-broker/broker"
	"log"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

func main() {
	log.Println("starting broker...")
	s := broker.New()
	s.Start()
}
