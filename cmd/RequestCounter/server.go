package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/RoanBrand/RequestCounter/internal/db"
	"github.com/pkg/errors"
)

type Server struct {
	ctx      context.Context
	s        http.Server
	db       *db.DB
	hostName string
}

func (s *Server) Init(ctx context.Context, listenAddr, dbFilePath string) {
	hostName, err := os.Hostname()
	if err != nil {
		log.Println("could not resolve hostname:", err.Error())
		// continue as not critical
	} else {
		s.hostName = hostName
	}

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
	ctx := r.Context()

	newClusterCount, err := s.makeClusterRequest(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		err := errors.WithMessage(err, "failed to contact cluster")
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newNodeCount := s.db.IncCount()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	_, err = fmt.Fprintf(
		w,
		"You are talking to instance %s%s.\nThis is request %d to this instance and request %d to the cluster.\n",
		s.hostName,
		s.s.Addr,
		newNodeCount,
		newClusterCount,
	)
	if err != nil {
		log.Println("error sending response:", err)
	}
}

// makeClusterRequest makes a request to cluster and
// returns new count of total requests made to it.
func (s *Server) makeClusterRequest(ctx context.Context) (uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
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
