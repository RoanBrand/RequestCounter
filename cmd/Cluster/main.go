package main

import (
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
)

var count uint64

func main() {
	s := http.Server{Addr: ":8024"}
	defer s.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		newCount := atomic.AddUint64(&count, 1)

		w.Write([]byte(strconv.FormatUint(newCount, 10)))
	})

	log.Println(s.ListenAndServe())
}
