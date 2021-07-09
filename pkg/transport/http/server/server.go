package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

type Server struct {
	httpServer *http.Server
	opts       *options
}

func New(opts ...Option) *Server {
	options := evaluateOptions(opts)
	s := &Server{
		opts: options,
	}

	s.httpServer = &http.Server{
		Handler:      options.handler,
		ReadTimeout:  options.timeouts.Read,
		WriteTimeout: options.timeouts.Write,
		IdleTimeout:  options.timeouts.Idle,
	}

	return s
}

// Start starts new http server.
// It's also listens for given context to be done, and if so it shutdowns
// server with configured timeout.
func (s *Server) Start(ctx context.Context) chan error {
	done := make(chan error)
	go func() {
		<-ctx.Done()
		shutdownContext, _ := context.WithTimeout(context.Background(), s.opts.timeouts.Shutdown)
		s.opts.logger.Info("shutting down server")
		err := s.httpServer.Shutdown(shutdownContext)
		if err != nil {
			done <- fmt.Errorf("(%s) shutdown HTTP server: %w", s.opts.name, err)
			return
		}

		close(done)
	}()
	go func() {
		if err := validate(s); err != nil {
			done <- fmt.Errorf("(%s) invalid server configuration: %w", s.opts.name, err)
		}

		listener, err := net.Listen("tcp", s.opts.listenAddr)
		if err != nil {
			done <- fmt.Errorf("(%s): %s", s.opts.name, err)
			return
		}

		s.opts.logger.Infof("running HTTP listener on %s", s.opts.listenAddr)
		if err := s.httpServer.Serve(listener); err != http.ErrServerClosed {
			done <- fmt.Errorf("(%s) serve HTTP: %w", s.opts.name, err)
			return
		}
	}()

	return done
}

func validate(s *Server) error {
	if s.httpServer.Handler == nil {
		return fmt.Errorf("(%s) no handler provided", s.opts.name)
	}

	return nil
}
