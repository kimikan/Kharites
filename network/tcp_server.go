package network

/*
 * Written by kimi kan, 2016-10
 * This file is used to wrapper the tcp server.
 */

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// ConnSet ...
type ConnSet map[net.Conn]struct{}

//TCPServer ...
type TCPServer struct {
	Addr  string
	ln    net.Listener
	conns ConnSet

	mutexConns sync.Mutex
	wgConns    sync.WaitGroup

	ConnHandler func(net.Conn) bool
}

//NewTCPServer ...
func NewTCPServer(addr string, handler func(net.Conn) bool) *TCPServer {
	p := &TCPServer{
		Addr:        addr,
		ConnHandler: handler,
	}

	return p
}

//Start ...
func (p *TCPServer) Start() bool {
	ln, err := net.Listen("tcp", p.Addr)
	if err != nil {
		fmt.Println("Start listen failed: ", err, p.Addr)
		return false
	}

	p.ln = ln
	p.conns = make(ConnSet)

	return true
}

// Run ...Single thread run.
func (p *TCPServer) Run() {
	//Consumer
	var tempDelay time.Duration
	for {
		conn, err := p.ln.Accept()

		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				fmt.Printf("accept error: %v; retrying in %v\n", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			fmt.Println(err, "Accept")
			return
		}
		tempDelay = 0

		p.mutexConns.Lock()
		p.conns[conn] = struct{}{}
		p.mutexConns.Unlock()
		p.wgConns.Add(1) //barrier

		go func() {
			if p.ConnHandler != nil {
				p.ConnHandler(conn)
			}
			conn.Close()
			p.mutexConns.Lock()
			delete(p.conns, conn)
			p.mutexConns.Unlock()
			p.wgConns.Done()
		}()
	}
}

//Stop ...
func (p *TCPServer) Stop() {
	fmt.Println("Stop...")
	p.ln.Close()
	p.mutexConns.Lock()
	for conn := range p.conns {
		conn.Close()
	}
	p.conns = nil
	p.mutexConns.Unlock()
	fmt.Println("Stop2...")
	p.wgConns.Wait()
	fmt.Println("Stop3...")
}
