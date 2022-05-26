package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var clusterAddr = os.Getenv("CLUSTER_ADDR")

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var s Server
	s.Init(ctx, os.Getenv("LISTEN_ADDR"), os.Getenv("DB_FILE"))
	defer s.Close()

	if err := s.Run(); err != nil {
		log.Println("server error:", err)
	}
}
