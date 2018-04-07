package ident

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/nasa9084/ident/infra"
	"github.com/nasa9084/syg"
)

// Server is a main application object.
type Server struct {
	server *http.Server
	closed chan struct{}
}

// NewServer returns a new server.
func NewServer(addr string, privKeyPath string, cfg infra.Config) (*Server, error) {
	env, err := infra.NewEnvironment(cfg, privKeyPath)
	if err != nil {
		return nil, err
	}
	router := mux.NewRouter()
	bindRoutes(router, env)

	s := &Server{
		server: &http.Server{
			Addr:    addr,
			Handler: router,
		},
		closed: make(chan struct{}),
	}
	return s, nil
}

// Run the server.
func (s *Server) Run() error {
	cancel := syg.Listen(s.Shutdown, os.Interrupt)
	defer cancel()

	log.Printf("server is listening on: %s", s.server.Addr)
	err := s.server.ListenAndServe()
	<-s.closed
	return err
}

// Shutdown shuts down the server gracefully.
func (s *Server) Shutdown(os.Signal) {
	defer close(s.closed)
	log.Print("server shutdown")

	s.server.Shutdown(context.Background())
}
