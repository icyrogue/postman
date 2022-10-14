package asyncstorage

import (
	"context"
	"log"
	"time"
)

type asyncStorage struct {
	data    [][]interface{}
	input   chan []interface{}
	storage Storage
	Options *Options
}

type Options struct {
	MaxBufferLength int
	MaxWaitTime     int
}

type Storage interface {
	Add(ctx context.Context, data [][]interface{}) error
}

func New(storage Storage) *asyncStorage {
	return &asyncStorage{input: make(chan []interface{}, 5),
		storage: storage}
}

func (as *asyncStorage) Start(ctx context.Context) {
	d := time.Duration(time.Second * time.Duration(as.Options.MaxWaitTime))
	timer := time.AfterFunc(d, as.append)
	defer timer.Stop()
	go func() {
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case v := <-as.input:
				timer.Reset(d)
				as.data = append(as.data, v)

				if len(as.data) == as.Options.MaxBufferLength {
					as.append()
				}
			}
		}
	}()
}

func (as *asyncStorage) GetInput() chan []interface{} {
	return as.input
}

func (as *asyncStorage) append() {
	if len(as.data) > 0 {
		log.Println("appending to db")
		err := as.storage.Add(context.Background(), as.data)
		if err != nil {
			log.Println(err.Error())
		}
		as.data = [][]interface{}{}
	}
}
