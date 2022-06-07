package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/RoanBrand/RequestCounter/internal/db"
)

func TestRequestHandler(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	os.Remove("test.test") // in case previous run failed

	s := Server{
		ctx: ctx,
		db:  db.NewDB("test.test"),
	}
	defer os.Remove("test.test")
	defer s.db.Close()

	num := 1000
	expected := make(map[uint64]struct{}, num)
	for i := 1; i <= num; i++ {
		expected[uint64(i)] = struct{}{}
	}
	results := make(chan uint64, num)

	var wg sync.WaitGroup
	wg.Add(num)
	errs := make(chan error, 1)

	for i := 0; i < num; i++ {
		go func() {
			defer wg.Done()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			s.requestHandler(w, req)
			res := w.Result()
			defer res.Body.Close()

			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				errs <- err
				return
			}

			if len(data) != 8 {
				errs <- fmt.Errorf("not 8 bytes. Is: %d", len(data))
				return
			}

			results <- binary.LittleEndian.Uint64(data)
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	for {
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case err := <-errs:
			t.Fatal(err)
		case res := <-results:
			if _, ok := expected[res]; !ok {
				t.Fatalf("already got %d", res)
			} else {
				delete(expected, res)
			}
		case <-done:
			// drain any remaining results
			for len(results) > 0 {
				res := <-results
				if _, ok := expected[res]; !ok {
					t.Fatalf("already got %d", res)
				} else {
					delete(expected, res)
				}
			}

			if len(expected) != 0 {
				t.Fatalf("didn't get all results. Left: %v", expected)
			}
			return
		}
	}
}
