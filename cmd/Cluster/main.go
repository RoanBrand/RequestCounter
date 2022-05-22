package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	s := http.Server{Addr: os.Getenv("LISTEN_ADDR")}
	defer s.Close()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-c
		log.Println("exiting:", sig)
		if err := stopServer(&s); err != nil {
			log.Println("error stopping server:", err)
		}
	}()

	if err := runServer(&s); err != nil {
		log.Println("server error:", err)
	}
}

func runServer(s *http.Server) error {
	var count uint64

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		newCount := atomic.AddUint64(&count, 1)
		resp := []byte(strconv.FormatUint(newCount, 10))

		if _, err := w.Write(resp); err != nil {
			log.Println("error", err)
		}
	})

	err := s.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}

	return err
}

func stopServer(s *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	return s.Shutdown(ctx)
}
