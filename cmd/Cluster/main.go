package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
)

func main() {
	s := http.Server{Addr: ":" + os.Getenv("LISTEN_PORT")}
	defer s.Close()

	var count uint64

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		newCount := atomic.AddUint64(&count, 1)
		resp := []byte(strconv.FormatUint(newCount, 10))

		if _, err := w.Write(resp); err != nil {
			log.Println("error", err)
		}
	})

	log.Println(s.ListenAndServe())
}
