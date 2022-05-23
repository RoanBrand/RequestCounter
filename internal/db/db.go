package db

import (
	"log"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/pkg/errors"
)

type db struct {
	count uint64
	flush chan struct{}
	file  string
}

func NewDB(dbFilePath string) *db {
	d := &db{
		flush: make(chan struct{}, 1),
		file:  dbFilePath,
	}

	// async db saver
	go func(d *db) {
		for range d.flush {
			if err := d.saveCount(); err != nil {
				log.Println("error persisting to disk:", err)
			}
		}
	}(d)

	if err := d.loadCount(); err != nil {
		log.Println("error loading saved value:", err)
	}

	return d
}

func (d *db) Close() error {
	close(d.flush)
	return nil
}

func (d *db) IncCount() uint64 {
	newCount := atomic.AddUint64(&d.count, 1)
	d.saveCountLazy()
	return newCount
}

func (d *db) saveCountLazy() {
	if len(d.flush) == 0 {
		select {
		case d.flush <- struct{}{}:
		default:
		}
	}
}

func (d *db) loadCount() error {
	fb, err := os.ReadFile(d.file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Wrap(err, "unable to read "+d.file)
	}

	if len(fb) > 8 {
		log.Println(d.file, "corrupted. Ignoring")
		return nil
	}

	d.count, err = strconv.ParseUint(string(fb), 10, 64)
	return err
}

func (d *db) saveCount() error {
	c := atomic.LoadUint64(&d.count)
	fb := strconv.FormatUint(c, 10)

	return os.WriteFile(d.file, []byte(fb), 0644)
}
