package gopipe

import (
	"fmt"
	"time"
)

type process func(interface{}) (interface{}, error)

// Pipeline is the definition of procedures
type Pipeline struct {
	steps   []step
	OutErr  chan PipeError
	outErr  chan PipeError
	running bool
	stop    chan struct{}
}

// Options is for passing params to pp
type Options struct {
	MaxRunning int
	InputLen   int
}

// Start pipeline procedure
func (p *Pipeline) Start(op Options) (chan interface{}, chan interface{}) {
	p.running = true
	p.OutErr = make(chan PipeError, op.InputLen)
	p.outErr = make(chan PipeError, op.InputLen)
	p.stop = make(chan struct{}, 0)
	inpChan := make(chan interface{}, op.InputLen)
	retChan := inpChan
	for i := 0; i < len(p.steps); i++ {
		if p.steps[i].maxRunning == 0 {
			p.steps[i].maxRunning = op.MaxRunning
		}
		outChan := make(chan interface{}, op.InputLen)
		p.steps[i].input = inpChan
		p.steps[i].output = outChan
		p.steps[i].outErr = p.OutErr
		p.steps[i].stop = p.stop
		inpChan = outChan
	}

	go p.readErrors()
	return retChan, inpChan
}

// AddFn will add a function to the pipeline
func (p *Pipeline) AddFn(
	name string,
	fn process,
	op Options,
) error {
	if p.running {
		return fmt.Errorf(
			"Can't add a function while pipeline is running",
		)
	}
	if p.steps == nil {
		p.steps = make([]step, 0)
	}
	p.steps = append(p.steps, step{
		name:       name,
		fn:         fn,
		maxRunning: op.MaxRunning,
	})
	return nil
}

// Stats will return pipeline status
func (p *Pipeline) Stats() []Stat {
	sts := make([]Stat, len(p.steps))
	for i := 0; i < len(p.steps); i++ {
		sts[i] = p.steps[i].mon.get()
	}
	return sts
}

// To avoid non reading clients
func (p *Pipeline) readErrors() {
	for {
		e := <-p.outErr
		select {
		case p.OutErr <- e:
		case <-time.After(1 * time.Second):
			// Clean err channels
			for len(p.OutErr) > 0 {
				<-p.OutErr
			}
		}
	}
}
