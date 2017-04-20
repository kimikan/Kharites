package dispatcher

/*
 * Written by kimi kan, 2016-10
 * This file not used currently, firstly i want to import a Dispatcher to
 * summerize the send request(also called response) with chanel
 * since the socket.write() is thread-safe,
 * but it still can be used for other operations.
 */

import (
	"Kharites/network"
	"runtime"
	"sync"
)

//Dispatcher ...
type Dispatcher struct {
	SendQueue chan interface{}
	IOQueue   chan interface{}

	Reader network.NetReader
	Writer network.NetWriter

	closeWg sync.WaitGroup
}

//NewDispatcher ...
func NewDispatcher(r network.NetReader, w network.NetWriter) *Dispatcher {
	d := &Dispatcher{
		Reader: r,
		Writer: w,
	}

	d.SendQueue = make(chan interface{})
	d.IOQueue = make(chan interface{})

	return d
}

//Start ...f means Taskhandler
func (d *Dispatcher) Start(f func(p interface{})) {
	go d.handleSendQueue()

	for i := 0; i < runtime.NumCPU(); i++ {
		go d.doWork(f)
	}
}

//Close ...
func (d *Dispatcher) Close() {
	//d.SendQueue
	close(d.SendQueue)
	close(d.IOQueue)
	d.closeWg.Wait()
}

//DoWork Multi-threads enabled
func (d *Dispatcher) doWork(f func(p interface{})) {
	d.closeWg.Add(1)
	defer d.closeWg.Done()

	for {
		s, err := <-d.IOQueue
		if err {
			break
		}
		f(s) //Real handler
	}
}

//EnqueueSendTask ...
func (d *Dispatcher) EnqueueSendTask(msg interface{}) {
	d.SendQueue <- msg
}

func (d *Dispatcher) handleSendQueue() {
	d.closeWg.Add(1)
	defer d.closeWg.Done()
	defer close(d.SendQueue)

	for {
		select {
		case s, err := <-d.SendQueue:
			if err {
				return
			}
			if s != nil {
				if !d.Writer.WriteMsg(s) {
					return
				}
			}
		} //end select
	} // end for
}
