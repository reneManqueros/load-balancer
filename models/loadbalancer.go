package models

import (
	"errors"
	"fmt"
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
	Network    string
	Source     string
	Backends   []Backend `yaml:"backends"`
	Mutex      sync.Mutex
	ConfigFile string
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
	ln, err := net.Listen(l.Network, l.Source)
	if err != nil {
		log.Fatalln(err)
	}

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

		log.Println(fmt.Sprintf(`routing from %v through %s`, sourceConnection.LocalAddr(), destination))
		go l.handleRequest(sourceConnection, destination.Address)
	}
}

func (l *LoadBalancer) handleRequest(sourceConnection net.Conn, destinationAddress string) {
	var destinationConnection net.Conn
	var err error

	destinationConnection, err = net.DialTimeout(l.Network, destinationAddress, 300*time.Millisecond)
	if err != nil {
		log.Println("handle request error", err)
		return
	}

	go copyIO(sourceConnection, destinationConnection)
	go copyIO(destinationConnection, sourceConnection)
}

func copyIO(src, dest net.Conn) {
	defer src.Close()
	defer dest.Close()
	io.Copy(src, dest)
}
