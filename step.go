package gopipe

import "fmt"

type step struct {
	fn         process
	maxRunning int
	name       string
	mon        monitor
	outErr     chan PipeError
	input      chan interface{}
	output     chan interface{}
	stop       chan struct{}
}

func (s *step) start() error {
	if s.maxRunning == 0 {
		// s.outErr <- PipeError{
		e := PipeError{
			Step: s.name,
			Err:  fmt.Errorf("Can't start 0 instances"),
		}
		return e
	}

	go s.mon.collect()

	available := make(chan struct{}, s.maxRunning)
	for i := 0; i < s.maxRunning; i++ {
		available <- struct{}{}
	}

	goon := true
	go func() {
		for goon {
			select {
			case v := <-s.input:
				<-available
				go func() {
					defer func() {
						available <- struct{}{}
					}()
					r, e := s.fn(v)
					s.output <- r
					s.outErr <- PipeError{
						Step: s.name,
						Err:  e,
					}
					if e != nil {
						s.mon.err()
					} else {
						s.mon.ok()
					}
				}()
			case <-s.stop:
				goon = false
			}
		}
	}()

	return nil
}

func (s *step) stats() Stat {
	return s.mon.get()
}
