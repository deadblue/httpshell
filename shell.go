package httpshell

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

type Shell struct {
	// Core HTTP handler
	core http.Handler
	// Listen network
	ln string
	// Listen address
	la string
	// HTTP server
	hs *http.Server

	cf int32
	ec chan error
}

func (s *Shell) die(err error) {
	if atomic.CompareAndSwapInt32(&s.cf, 0, 1) {
		if err != nil {
			s.ec <- err
		}
		close(s.ec)
	}
}

func (s *Shell) Start() (err error) {
	l, err := net.Listen(s.ln, s.la)
	if err != nil {
		return
	}
	if ul, ok := l.(*net.UnixListener); ok {
		ul.SetUnlinkOnClose(true)
	}
	if ob, ok := s.core.(StartupObserver); ok {
		if err = ob.BeforeStartup(); err != nil {
			l.Close()
			return
		}
	}
	go func(l net.Listener) {
		log.Printf("Start HTTP server at %s://%s", s.ln, s.la)
		err = s.hs.Serve(l)
		if !errors.Is(err, http.ErrServerClosed) {
			s.die(err)
		}
	}(l)
	return
}

func (s *Shell) Stop() {
	go func() {
		if s.cf == 0 {
			err := s.hs.Shutdown(context.Background())
			if ob, ok := s.core.(ShutdownObserver); ok {
				ob.AfterShutdown()
			}
			s.die(err)
		}
	}()
}

func (s *Shell) Error() <-chan error {
	return s.ec
}

func (s *Shell) Run(stopSigs ...os.Signal) (err error) {
	if len(stopSigs) == 0 {
		stopSigs = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}
	// Handle OS signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, stopSigs...)
	defer signal.Stop(sigCh)
	// Start server
	if err = s.Start(); err != nil {
		return
	}
	// Handle event
	for running := true; running; {
		select {
		case <-sigCh:
			log.Println("Shutting down server ...")
			s.Stop()
		case err = <-s.Error():
			running = false
		}
	}
	log.Println("Server is down, byebye!")
	return
}

func New(network, addr string, handler http.Handler) *Shell {
	return &Shell{
		ln:   network,
		la:   addr,
		core: handler,
		hs: &http.Server{
			Handler: handler,
		},
		ec: make(chan error),
	}
}
