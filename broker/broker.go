package broker

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/cloudfoundry-community/types-cf"
)

// Broker represents a running CF Service Broker API
type Broker struct {
	router *router
}

// New creates a loaded instance o the broker
func New(p cf.ServiceProvider) (*Broker, error) {
	return &Broker{
		router: newRouter(newHandler(p)),
	}, nil
}

// Start the broker
func (b *Broker) Start() {

	addr := fmt.Sprintf("%s:%d", Config.CFEnv.Host, Config.CFEnv.Port)
	log.Printf("starting: %s", addr)

	sigCh := make(chan os.Signal, 1)

	// make sure we can shutdown gracefully
	signal.Notify(sigCh, os.Interrupt)

	errCh := make(chan error, 1)

	go func() {
		errCh <- http.ListenAndServe(addr, b.router)
	}()

	// non blocking as some of these cf ops are kind of lengthy
	select {
	case err := <-errCh:
		log.Printf("broker error: %v", err)
	case sig := <-sigCh:
		var _ = sig
		log.Print("broker done")
	}

}
