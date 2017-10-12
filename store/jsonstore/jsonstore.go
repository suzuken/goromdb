package jsonstore

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/yowcow/go-romdb/store"
)

type Data map[string]string

type Store struct {
	file   string
	data   Data
	logger *log.Logger

	dataNodeQuit chan bool
	dataNodeWg   *sync.WaitGroup

	watcherQuit chan bool
	watcherWg   *sync.WaitGroup
}

func New(file string) (store.Store, error) {
	var data Data
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	dataUpdate := make(chan Data)

	dataNodeQuit := make(chan bool)
	dataNodeWg := &sync.WaitGroup{}

	watcherQuit := make(chan bool)
	watcherWg := &sync.WaitGroup{}

	s := &Store{file, data, logger, dataNodeQuit, dataNodeWg, watcherQuit, watcherWg}

	boot := make(chan bool)

	dataNodeWg.Add(1)
	go s.startDataNode(boot, dataUpdate)
	<-boot

	watcherWg.Add(1)
	go s.startWatcher(boot, dataUpdate)
	<-boot

	close(boot)

	return s, nil
}

func (s *Store) startDataNode(boot chan<- bool, dataIn <-chan Data) {
	defer s.dataNodeWg.Done()

	if data, err := LoadJSON(s.file); err == nil {
		s.data = data
	}

	boot <- true
	s.logger.Print("-> datanode started!")

	for {
		select {
		case data := <-dataIn:
			s.logger.Print("-> datanode updated!")
			s.data = data
		case <-s.dataNodeQuit:
			s.logger.Print("-> datanode finished!")
			return
		}
	}
}

func (s Store) startWatcher(boot chan<- bool, dataOut chan<- Data) {
	defer s.watcherWg.Done()

	var lastModified time.Time

	if fi, err := os.Stat(s.file); err == nil {
		lastModified = fi.ModTime()
	}

	d := 5 * time.Second
	t := time.NewTimer(d)

	boot <- true
	s.logger.Print("-> watcher started!")

	for {
		select {
		case <-t.C:
			if fi, err := os.Stat(s.file); err == nil {
				if fi.ModTime() != lastModified {
					lastModified = fi.ModTime()
					if data, err := LoadJSON(s.file); err == nil {
						dataOut <- data
					} else {
						s.logger.Print("-> watcher failed reading data from file: ", err)
					}
				}
			} else {
				s.logger.Print("-> watcher file check failed: ", err)
			}
			t.Reset(d)
		case <-s.watcherQuit:
			if !t.Stop() {
				<-t.C
			}
			s.logger.Print("-> watcher finished!")
			return
		}
	}
}

func (s Store) Get(key string) (string, error) {
	if v, ok := s.data[key]; ok {
		return v, nil
	}
	return "", store.KeyNotFoundError(key)
}

func (s Store) Shutdown() error {
	s.watcherQuit <- true
	close(s.watcherQuit)
	s.watcherWg.Wait()

	s.dataNodeQuit <- true
	close(s.dataNodeQuit)
	s.dataNodeWg.Wait()

	return nil
}

func LoadJSON(file string) (Data, error) {
	var data Data

	b, err := ioutil.ReadFile(file)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(b, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}
