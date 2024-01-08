package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
)

type server struct {
	Address string
}

func New(address string) *server {
	return &server{
		Address: address,
	}
}

func (s *server) Start() {
	log.Printf("Listening at http://%s", s.Address)

	httpServer := http.Server{
		Addr: fmt.Sprintf(":%s", s.Address),
	}

	idleConnectionsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP Server Shutdown Error: %v", err)
		}
		close(idleConnectionsClosed)
	}()

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe Error: %v", err)
	}

	<-idleConnectionsClosed

	log.Printf("Bye bye")
}
