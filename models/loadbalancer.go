package models

import (
	"errors"
	"fmt"
	"github.com/remeh/sizedwaitgroup"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

type LoadBalancer struct {
	Network     string
	Source      string
	ConfigFile  string
	Backends    []Backend `yaml:"backends"`
	Mutex       sync.Mutex
	IsVerbose   bool
	Timeout     int
	Connections int
}

func (l *LoadBalancer) GetBackend() (b Backend, err error) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	if len(l.Backends) == 0 {
		err = errors.New("no backends available")
		return
	}
	randomBackendIndex := rand.Intn(len(l.Backends))
	b = l.Backends[randomBackendIndex]
	return
}

func (l *LoadBalancer) Exists(backend Backend) bool {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	for _, b := range l.Backends {
		if b.Address == backend.Address {
			return true
		}
	}
	return false
}
func (l *LoadBalancer) ToDisk() {
	b, err := yaml.Marshal(&l.Backends)
	if err != nil {
		log.Println(err)
	}

	err = ioutil.WriteFile(l.ConfigFile, b, 0644)
	if err != nil {
		log.Fatalln("error writing backends file", err)
	}
}

func (l *LoadBalancer) FromDisk() {
	if _, err := os.Stat(l.ConfigFile); errors.Is(err, os.ErrNotExist) {
		log.Println("no backends file, creating blank")
		_, err = os.Create(l.ConfigFile)
		if err != nil {
			log.Fatalln("couldnt create backends file")
		}
		return
	}

	b, err := ioutil.ReadFile(l.ConfigFile)
	if err != nil {
		log.Fatalln("error loading backends file", err)
		return
	}
	var backends []Backend
	_ = yaml.Unmarshal(b, &backends)
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	l.Backends = backends

}

func (l *LoadBalancer) Add(backend Backend) bool {
	if l.Exists(backend) == false {
		l.Mutex.Lock()
		defer l.Mutex.Unlock()
		l.Backends = append(l.Backends, backend)
		l.ToDisk()
		return true
	}
	return false
}

func (l *LoadBalancer) Remove(backend Backend) bool {
	if l.Exists(backend) == true {
		l.Mutex.Lock()
		defer l.Mutex.Unlock()
		var tmpBackends []Backend
		for _, b := range l.Backends {
			if b.Address != backend.Address {
				tmpBackends = append(tmpBackends, b)
			}
		}
		l.Backends = tmpBackends
		l.ToDisk()
		return true
	}
	return false
}

func (l *LoadBalancer) Listen() {
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			log.Println("UNHANDLED ERROR! :", err)
		}
	}()

	ln, err := net.Listen(l.Network, l.Source)
	if err != nil {
		log.Println(err)
	}

	swg := sizedwaitgroup.New(l.Connections)
	for {
		sourceConnection, err := ln.Accept()
		if err != nil {
			panic(err)
		}

		destination, err := l.GetBackend()
		if err != nil {
			log.Println(err)
			continue
		}

		if l.IsVerbose == true {
			log.Println(fmt.Sprintf(`routing from %v through %s`, sourceConnection.LocalAddr(), destination))
		}
		swg.Add()
		go l.handleRequest(sourceConnection, destination.Address, &swg)
	}
}

func (l *LoadBalancer) handleRequest(sourceConnection net.Conn, destinationAddress string, tswg *sizedwaitgroup.SizedWaitGroup) {
	var destinationConnection net.Conn
	var err error

	destinationConnection, err = net.DialTimeout(l.Network, destinationAddress, time.Duration(l.Timeout)*time.Millisecond)
	if err != nil {
		log.Println("handle request error", err)
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go copyIO(sourceConnection, destinationConnection, &wg)
	go copyIO(destinationConnection, sourceConnection, &wg)
	wg.Wait()
	tswg.Done()
}

func copyIO(src, dest net.Conn, twg *sync.WaitGroup) {
	defer src.Close()
	defer dest.Close()
	io.Copy(src, dest)
	defer twg.Done()
}
