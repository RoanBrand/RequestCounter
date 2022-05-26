package db_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

// Justification for using len(chan):
/*
	Reading a channel's len is not a data race by itself,
	but the value returned is racy and generally cannot be relied upon.

	But in this specific case its usage is fine I believe, because
	in both cases of 0 and 1, we can still guarantee that the new
	count will get saved afterwards.

	The inspiration for my usage of it comes from this article from a NATS engineer:
	https://www.oreilly.com/content/scaling-messaging-in-go-network-clients/

	I can also see it still being used in this way in nats.go. Please see:
	https://github.com/nats-io/nats.go/blob/144a3b25a04c2dff2657c24b49652f9b1e652daf/nats.go#L3526

	On my machine BenchmarkWithLen is faster than BenchmarkWithoutLen.
	I believe this is due to an optimization where len(chan) is efficient because it does
	not make use of synchronization, so in the case where the buffered chan is already filled,
	we can skip the select with attempt chan send, which does make use of sync code.
	Please see:
	https://groups.google.com/g/golang-nuts/c/yQw1Wx6BoUU
	https://groups.google.com/g/golang-nuts/c/L0wIBDr3HCc
*/

const workers = 10

func BenchmarkWithLen(b *testing.B) {
	var count, lastSaved, skipped uint64
	flush := make(chan struct{}, 1)
	var wg sync.WaitGroup
	wg.Add(workers)

	b.ReportAllocs()
	fmt.Println()
	b.ResetTimer()

	go func() {
		for i := 0; i < workers; i++ {
			go func() {
				for i := 0; i < b.N; i++ {
					atomic.AddUint64(&count, 1)
					if len(flush) == 0 {
						select {
						case flush <- struct{}{}:
						default:
						}
					}
				}
				wg.Done()
			}()
		}

		wg.Wait()
		close(flush)
	}()

	for range flush {
		newCount := atomic.LoadUint64(&count)
		if newCount < lastSaved {
			b.Fatal(newCount)
		}

		if newCount > lastSaved {
			skipped += newCount - lastSaved - 1
		}
		lastSaved = newCount
	}

	b.StopTimer()

	total := uint64(b.N) * workers
	if lastSaved != total {
		b.Fatal(lastSaved, total)
	}

	per := float64(skipped*100) / float64(total)
	fmt.Println("skipped", skipped, "saves out of", total, per, "%")
}

func BenchmarkWithoutLen(b *testing.B) {
	var count, lastSaved, skipped uint64
	flush := make(chan struct{}, 1)
	var wg sync.WaitGroup
	wg.Add(workers)

	b.ReportAllocs()
	fmt.Println()
	b.ResetTimer()

	go func() {
		for i := 0; i < workers; i++ {
			go func() {
				for i := 0; i < b.N; i++ {
					atomic.AddUint64(&count, 1)
					select {
					case flush <- struct{}{}:
					default:
					}
				}
				wg.Done()
			}()
		}

		wg.Wait()
		close(flush)
	}()

	for range flush {
		newCount := atomic.LoadUint64(&count)
		if newCount < lastSaved {
			b.Fatal(newCount)
		}

		if newCount > lastSaved {
			skipped += newCount - lastSaved - 1
		}
		lastSaved = newCount
	}

	b.StopTimer()

	total := uint64(b.N) * workers
	if lastSaved != total {
		b.Fatal(lastSaved, total)
	}

	per := float64(skipped*100) / float64(total)
	fmt.Println("skipped", skipped, "saves out of", total, per, "%")
}
