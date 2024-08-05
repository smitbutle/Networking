package main

// Demonstrates how to create a simple HTTP server using raw syscalls. 
// This is a simple HTTP server that listens on port 8080 and logs the incoming requests to a file called log.txt.
// The server responds with a simple HTML response "hello world" to every incoming request.

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/textproto"
	"os"
	"syscall"
)

type netSocket struct {
	fd int
}

func (ns netSocket) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	n, err := syscall.Read(ns.fd, p)
	if err != nil {
		n = 0
	}
	return n, err
}

func (ns netSocket) Write(p []byte) (int, error) {
	n, err := syscall.Write(ns.fd, p)
	if err != nil {
		n = 0
	}
	return n, err
}

// Creates a new netSocket for the next pending connection request.
func (ns *netSocket) Accept() (*netSocket, error) {
	nfd, _, err := syscall.Accept(ns.fd)
	if err == nil {
		syscall.CloseOnExec(nfd)
	}
	if err != nil {
		return nil, err
	}
	return &netSocket{nfd}, nil
}

func (ns *netSocket) Close() error {
	return syscall.Close(ns.fd)
}

// Creates a new socket file descriptor, binds it and listens on it.
func newNetSocket(ip net.IP, port int) (*netSocket, error) {

	syscall.ForkLock.Lock()

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, os.NewSyscallError("socket", err)
	}
	syscall.ForkLock.Unlock()

	if err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		syscall.Close(fd)
		return nil, os.NewSyscallError("setsockopt", err)
	}

	// Bind the socket to a port
	sa := &syscall.SockaddrInet4{Port: port}
	copy(sa.Addr[:], ip)
	if err = syscall.Bind(fd, sa); err != nil {
		return nil, os.NewSyscallError("bind", err)
	}

	// Listen for incoming connections.
	if err = syscall.Listen(fd, syscall.SOMAXCONN); err != nil {
		return nil, os.NewSyscallError("listen", err)
	}

	return &netSocket{fd: fd}, nil
}

func main() {

	ip := net.ParseIP("127.0.0.1")
	port := 8080
	socket, err := newNetSocket(ip, port)
	if err != nil {
		panic(err)
	}
	defer socket.Close()

	log.Print("======================")
	log.Printf("Server Started! addr: http://%s:%d", ip, port)
	log.Print("======================")
	for {
		// Block until incoming connection
		rw, e := socket.Accept()
		log.Print()
		log.Print()
		log.Printf("Incoming connection")
		if e != nil {
			panic(e)
		}

		// Read request
		log.Print("Reading request")
		tp := textproto.NewReader(bufio.NewReader(*rw))

		// can offload writing log using goroutine, might try it later :)
		func() {
			for {
				s, err := tp.ReadLine()

				if s == "" || err != nil  {
					break
				}

				file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil {
					log.Print("[file opening error]" , err)
					break
				}

				defer file.Close()
				fmt.Fprintf(file, "%s\n", s)
				
			}
		}()

		log.Print("Writing response")
		io.WriteString(rw, 
			"HTTP/1.1 200 OK\r\n"+
			"Content-Type: text/html; charset=utf-8\r\n"+
			"Content-Length: 20\r\n"+
			"\r\n"+
			"<h1>hello world</h1>")
		rw.Close()
	}
}
