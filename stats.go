package gopipe

import (
	"sync"
	"time"
)

const defaultTimeout = 1

type monitor struct {
	mtx sync.Mutex

	okReader  chan struct{}
	errReader chan struct{}

	MaxRunning int
	Name       string
	Success    int
	Errors     int
}

func (m *monitor) get() Stat {
	m.mtx.Lock()
	s := Stat{
		MaxRunning: m.MaxRunning,
		Name:       m.Name,
		Success:    m.Success,
		Errors:     m.Errors,
		Time:       time.Now().UTC().Unix(),
	}
	m.Success = 0
	m.Errors = 0
	m.mtx.Unlock()
	return s
}

func (m *monitor) collect() {
	m.okReader = make(chan struct{}, m.MaxRunning)
	m.errReader = make(chan struct{}, m.MaxRunning)

	for {
		select {
		case <-m.okReader:
			m.mtx.Lock()
			m.Success++
			m.mtx.Unlock()
		case <-m.errReader:
			m.mtx.Lock()
			m.Errors++
			m.mtx.Unlock()
		}
	}
}

func (m *monitor) ok() {
	go func() {
		select {
		case m.okReader <- struct{}{}:
		case <-time.After(defaultTimeout * time.Minute):
		}
	}()
}

func (m *monitor) err() {
	go func() {
		select {
		case m.errReader <- struct{}{}:
		case <-time.After(defaultTimeout * time.Minute):
		}
	}()
}

// Stat will expose step stats
type Stat struct {
	MaxRunning int
	Name       string
	Success    int
	Errors     int
	Time       int64
}
