package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

const clusterAddr = "http://cluster:8084"

var count uint64

func main() {
	hostName, err := os.Hostname()
	if err != nil {
		log.Println("could not resolve hostname:", err.Error())
	}

	s := http.Server{Addr: ":8083"}
	defer s.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		newClusterCount, err := requestCluster()
		if err != nil {
			err := errors.WithMessage(err, "failed to contact cluster")
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		newNodeCount := atomic.AddUint64(&count, 1)

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(
			w,
			"You are talking to instance %s%s.\nThis is request %d to this instance and request %d to the cluster.\n",
			hostName,
			s.Addr,
			newNodeCount,
			newClusterCount,
		)
	})

	log.Println(s.ListenAndServe())
}

// requestCluster make request to cluster and
// returns new count of total requests made to cluster.
func requestCluster() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, clusterAddr, nil)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("cluster error: " + resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	bn, err := strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	return bn, nil
}
