package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/RoanBrand/RequestCounter/internal/db"

	"github.com/pkg/errors"
)

var clusterAddr = os.Getenv("CLUSTER_ADDR")
var data = db.NewDB(os.Getenv("DB_FILE"))

func main() {
	defer data.Close()

	hostName, err := os.Hostname()
	if err != nil {
		log.Println("could not resolve hostname:", err.Error())
		// continue as not critical
	}

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

	if err := runServer(&s, hostName); err != nil {
		log.Println("server error:", err)
	}
}

func runServer(s *http.Server, hostName string) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		newClusterCount, err := requestCluster()
		if err != nil {
			err := errors.WithMessage(err, "failed to contact cluster")
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		newNodeCount := data.IncCount()
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		_, err = fmt.Fprintf(
			w,
			"You are talking to instance %s%s.\nThis is request %d to this instance and request %d to the cluster.\n",
			hostName,
			s.Addr,
			newNodeCount,
			newClusterCount,
		)
		if err != nil {
			log.Println("error sending response:", err)
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
