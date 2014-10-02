package broker

import (
	"fmt"
	"github.com/intel-data/types-cf"
	"log"
	"net/http"
	"os"
	"os/signal"
)

type broker struct {
	router *router
}

// New creates a loaded instance o the broker
func New(p cf.ServiceProvider) (*broker, error) {
	return &broker{newRouter(newHandler(p))}, nil
}

// Start the broker
func (b *broker) Start() {

	sigCh := make(chan os.Signal, 1)

	// make sure we can shutdown gracefully
	signal.Notify(sigCh, os.Interrupt)

	errCh := make(chan error, 1)

	go func() {
		addr := fmt.Sprintf(":%v", Config.CFEnv.Port)
		log.Printf("broker started on%v", addr)
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
