package main

import (
	"log"
)

var c Config = Config{}

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	c.initialize()
}

func main() {
	log.Println("starting server...")
	s := &Server{config: c}
	s.start()
}
