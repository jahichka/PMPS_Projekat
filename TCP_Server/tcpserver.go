package main

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type TCPServer struct {
	Wg       *sync.WaitGroup
	ctx      context.Context
	logger   log.Entry
	listener *net.TCPListener
}

// NewTCPServer creates a simple interface for sending and recieving messages over TCP
func NewTCPServer(ctx context.Context, port string, logger log.Entry) (*TCPServer, error) {
	addr, err := net.ResolveTCPAddr("tcp", port)
	if err != nil {
		return nil, err
	}

	lsn, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &TCPServer{
		listener: lsn,
		logger:   logger,
		ctx:      ctx,
		Wg:       &sync.WaitGroup{},
	}, nil
}

// Start starts the server to listen for incoming connections until context is complete
func (srw *TCPServer) Start() {
	errChan := make(chan error)
	go func() {
		for {
			select {
			case err := <-errChan:
				srw.logger.Fatal(err.Error())
			case <-srw.ctx.Done():
				srw.Shutdown()
				return
			default:
				srw.logger.Info("Listening for new connection ... ")
				conn, err := srw.listener.Accept()
				if err != nil {
					fmt.Printf("no new clients\n")
					continue // if no connection was made, skip
				}
				srw.Wg.Add(1)
				go srw.connectionHandler(conn) // spawns a handler for connected client
			}
		}
	}()
}

// Shutdown initiates a clean shutdown routine for server
func (srw *TCPServer) Shutdown() {
	srw.listener.Close()
}

// connectionHandler reads from and writes data to a TCP client connected to the server
func (srw *TCPServer) connectionHandler(conn net.Conn) {

	// write anything to test the connection
	_, err := conn.Write([]byte("Welcome to my humble server\n"))
	if err != nil {
		srw.logger.Error(err.Error())
		return
	}
	srw.logger.Infof("Connection accepted: %v", conn.RemoteAddr().String())

	buffer := make([]byte, 1024)
loop:
	for {
		select {
		case <-srw.ctx.Done():
			break loop
		default:
			conn.SetDeadline(time.Now().Add(time.Minute)) // don't block if the connection is not alive
			size, err := conn.Read(buffer)
			if err != nil {
				break loop
			}
			logMsg := fmt.Sprintf("Received message from %s - '%s'", conn.RemoteAddr().String(), strings.TrimSuffix(string(buffer[:size]), "\n"))
			srw.logger.Info(logMsg)
			_, err = conn.Write([]byte("Your message has been received\n"))
			if err != nil {
				srw.logger.Errorf("Message could not be sent to %s", conn.RemoteAddr().String())
				break loop
			}
		}
	}
	conn.Close()
	srw.logger.Infof("Connection closed: %s", conn.RemoteAddr().String())
	srw.Wg.Done()
}
