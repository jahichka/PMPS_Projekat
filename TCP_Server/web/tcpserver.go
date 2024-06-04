package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var tcpServer *TCPServer

type TCPServer struct {
	Wg       *sync.WaitGroup
	ctx      context.Context
	logger   logrus.Entry
	listener *net.TCPListener
	stop     context.CancelFunc
	devices  map[string]*Device
	msgChan  chan *Message
	SendChan chan *Message
}

func CreateTCPServer(port string, logger logrus.Entry) error {
	ctx, stopServer := context.WithCancel(context.Background())
	server, err := newTCPServer(ctx, ":8080", *logger.WithFields(logrus.Fields{"prefix": "[TCP SERVER]"}))
	if err != nil {
		return err
	}
	tcpServer = server
	server.stop = stopServer
	server.Start()
	return nil
}

func StopTCPServer() {
	tcpServer.logger.Info(" Shutdown initiated")
	tcpServer.stop()
	tcpServer.Wg.Wait()
	tcpServer.logger.Info("\n-- -- Shutdown complete -- --\n")
}

// NewTCPServer creates a simple interface for sending and recieving messages over TCP
func newTCPServer(ctx context.Context, port string, logger logrus.Entry) (*TCPServer, error) {
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
		devices:  CreateMockDevices(),
	}, nil
}

// Start starts the server to listen for incoming connections until context is complete
func (srw *TCPServer) Start() {
	errChan := make(chan error)

	srw.msgChan = tcpServer.messageSender()

	go func() {
		for {
			select {
			case err := <-errChan:
				srw.logger.Fatal(err.Error())
			case <-srw.ctx.Done():
				srw.Shutdown()
				return
			default:
				srw.listener.SetDeadline(time.Now().Add(time.Second + 2))
				conn, err := srw.listener.Accept()
				if err != nil {
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

func (srw *TCPServer) CloseConn(conn net.Conn) {
	conn.Close()
	srw.logger.Infof("Connection closed: %s", conn.RemoteAddr().String())
	srw.Wg.Done()
}

func (srw *TCPServer) CloseDev(dev *Device) {
	srw.SendMessage(dev.ID, "connection closed")
	render := fmt.Sprintf(DEV_HTML_FULL, dev.ID, dev.Name, dev.ID, COLOR_OFF)
	WSMessage(dev.ID, EVENT_STATE, "logout", render)
	dev.State = STATE_OFF
}

// connectionHandler reads from and writes data to a TCP client connected to the server
func (srw *TCPServer) connectionHandler(conn net.Conn) {
	defer srw.CloseConn(conn)
	buffer := make([]byte, 1024)

	device, err := srw.AuthHandler(conn, buffer)
	if err != nil {
		return
	}
	defer srw.CloseDev(device)

	render := fmt.Sprintf(DEV_HTML_FULL, device.ID, device.Name, device.ID, COLOR_ON)
	WSMessage(device.ID, EVENT_STATE, "login", render)
	countdown := 0

loop:
	for {
		select {
		case <-srw.ctx.Done():
			break loop
		case msg := <-device.WriteChan:
			conn.SetDeadline(time.Now().Add(time.Second * 5))
			if _, err := conn.Write([]byte(msg)); err != nil {
				fmt.Println(err)
				break loop
			}
		default:
			conn.SetDeadline(time.Now().Add(time.Second)) // don't block if the connection is not alive
			size, err := conn.Read(buffer)
			if err != nil {
				if err != io.EOF && countdown < 360 {
					countdown++
					continue
				}
				break loop
			}
			countdown = 0
			srw.msgChan <- &Message{conn, device.ID, buffer, size}
		}
	}
}

func (srw *TCPServer) AuthHandler(conn net.Conn, buffer []byte) (*Device, error) {
	srw.SendMessage(conn.RemoteAddr().String(), "Auth request")

	_, err := conn.Write([]byte("auth"))
	if err != nil {
		srw.logger.Error(err.Error())
		return nil, err
	}

	conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	size, err := conn.Read(buffer)
	if err != nil {
		srw.SendMessage(conn.RemoteAddr().String(), "Auth timeout")
		return nil, err
	}

	var newDev Device
	if err := json.Unmarshal(buffer[:size], &newDev); err != nil {
		fmt.Println(string(buffer[:size]))
		srw.SendMessage("INTERNAL", "Json parsing error")
		return nil, err
	}

	if newDev.ID == "" {
		srw.SendMessage(conn.RemoteAddr().String(), fmt.Sprintf("Failed to autenticate (bad ID '%s')", newDev.ID))
		return nil, errors.New("Bad login id")
	}

	if _, devExists := srw.devices[newDev.ID]; devExists {
		if newDev.Auth == srw.devices[newDev.ID].Auth {
			srw.SendMessage(conn.RemoteAddr().String(), fmt.Sprintf("Authenticated as %s", newDev.ID))
		} else {
			srw.SendMessage(conn.RemoteAddr().String(), fmt.Sprintf("Failed to autenticate as %s (wrong auth token)", newDev.ID))
			return nil, err
		}
	} else {
		srw.SendMessage(conn.RemoteAddr().String(), fmt.Sprintf("New device register as %s", newDev.ID))
	}

	newDev.State = STATE_ON
	newDev.WriteChan = make(chan string, 4)
	srw.devices[newDev.ID] = &newDev

	return &newDev, nil
}
