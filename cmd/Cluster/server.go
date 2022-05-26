package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/RoanBrand/RequestCounter/internal/db"
)

type Server struct {
	ctx context.Context
	s   http.Server
	db  *db.DB
}

func (s *Server) Init(ctx context.Context, listenAddr, dbFilePath string) {
	s.ctx = ctx
	s.db = db.NewDB(dbFilePath)

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.requestHandler)
	s.s.Handler = mux

	s.s.Addr = listenAddr

	s.s.BaseContext = func(_ net.Listener) context.Context {
		return s.ctx
	}

	go func(s *Server) {
		<-s.ctx.Done()
		log.Println("stopping server")
		if err := s.Close(); err != nil {
			log.Println("error stopping server:", err)
		}
	}(s)
}

func (s *Server) Run() error {
	err := s.s.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}

	return err
}

func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := s.s.Shutdown(ctx)
	if err != nil {
		if err == http.ErrServerClosed {
			return nil
		}

		return err
	}

	return s.db.Close()
}

func (s *Server) requestHandler(w http.ResponseWriter, r *http.Request) {
	newCount := s.db.IncCount()
	resp := []byte(strconv.FormatUint(newCount, 10))

	if _, err := w.Write(resp); err != nil {
		log.Println("error", err)
	}
}
