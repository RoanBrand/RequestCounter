package db

import (
	"io/fs"
	"log"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/pkg/errors"
)

type DB struct {
	count uint64
	flush chan struct{}
	file  string
}

func NewDB(dbFilePath string) *DB {
	d := &DB{
		flush: make(chan struct{}, 1),
		file:  dbFilePath,
	}

	// async db flusher
	go func(d *DB) {
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

func (d *DB) Close() error {
	close(d.flush)
	return nil
}

func (d *DB) IncCount() uint64 {
	newCount := atomic.AddUint64(&d.count, 1)
	d.notifyFlusher()
	return newCount
}

func (d *DB) notifyFlusher() {
	if len(d.flush) == 0 {
		select {
		case d.flush <- struct{}{}:
		default:
		}
	}
}

var lastSave uint64

func (d *DB) saveCount() error {
	c := atomic.LoadUint64(&d.count)
	if c == lastSave {
		return nil
	}

	fb := strconv.FormatUint(c, 10)
	err := os.WriteFile(d.file, []byte(fb), 0644)
	if err != nil {
		return errors.Wrap(err, "unable to save "+d.file)
	}

	lastSave = c
	return nil
}

func (d *DB) loadCount() error {
	fb, err := os.ReadFile(d.file)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return errors.Wrap(err, "unable to read "+d.file)
	}

	if len(fb) > 8 {
		log.Println(d.file, d.file+" corrupted. Ignoring")
		return nil
	}

	d.count, err = strconv.ParseUint(string(fb), 10, 64)
	return err
}
