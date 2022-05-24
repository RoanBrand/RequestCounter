package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/RoanBrand/RequestCounter/internal/db"
)

var data = db.NewDB(os.Getenv("DB_FILE"))

func main() {
	defer data.Close()

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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		newCount := data.IncCount()
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
