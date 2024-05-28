package web

import (
	"context"
	"encoding/json"
	"fmt"
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

func (srw *TCPServer) CloseConn(conn net.Conn) {
	conn.Close()
	srw.logger.Infof("Connection closed: %s", conn.RemoteAddr().String())
	srw.Wg.Done()
}

func (srw *TCPServer) CloseDev(dev *Device) {
	srw.SendMessage(dev.ID, "connection closed")
	renderOff := fmt.Sprintf(DEV_HTML, dev.Name, dev.ID, COLOR_OFF)
	WSMessage(dev.ID, EVENT_STATE, "logout", renderOff)
	dev.State = STATE_OFF
}

// connectionHandler reads from and writes data to a TCP client connected to the server
func (srw *TCPServer) connectionHandler(conn net.Conn) {
	buffer := make([]byte, 1024)
	defer srw.CloseConn(conn)

	srw.SendMessage(conn.RemoteAddr().String(), "Auth request")

	// write anything to test the connection
	_, err := conn.Write([]byte("auth"))
	if err != nil {
		srw.logger.Error(err.Error())
		return
	}

	conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	size, err := conn.Read(buffer)
	if err != nil {
		srw.SendMessage(conn.RemoteAddr().String(), "Auth timeout")
		return
	}

	var jsonData Device
	if err := json.Unmarshal(buffer[:size], &jsonData); err != nil {
		fmt.Println(string(buffer[:size]))
		srw.SendMessage("INTERNAL", "Json parsing error")
	}

	if jsonData.ID == "" {
		srw.SendMessage(conn.RemoteAddr().String(), fmt.Sprintf("Failed to autenticate (bad ID '%s')", jsonData.ID))
		return
	}
	if _, devExists := srw.devices[jsonData.ID]; !devExists {
		srw.devices[jsonData.ID] = &jsonData
		srw.SendMessage(conn.RemoteAddr().String(), fmt.Sprintf("New device register as %s", jsonData.ID))
	} else {
		if jsonData.Auth == srw.devices[jsonData.ID].Auth {
			srw.SendMessage(conn.RemoteAddr().String(), fmt.Sprintf("Authenticated as %s", jsonData.ID))
		} else {
			srw.SendMessage(conn.RemoteAddr().String(), fmt.Sprintf("Failed to autenticate as %s (wrong auth token)", jsonData.ID))
			return
		}
	}
	srw.devices[jsonData.ID].State = STATE_ON

	render := fmt.Sprintf(DEV_HTML, jsonData.Name, jsonData.ID, COLOR_ON)
	WSMessage(jsonData.ID, EVENT_STATE, "login", render)

	defer srw.CloseDev(srw.devices[jsonData.ID])

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
			fmt.Println("WRITING A MESSAGE")
			srw.msgChan <- &Message{
				Conn:   conn,
				Dev:    srw.devices[jsonData.ID],
				Buffer: buffer,
				Size:   size,
			}
		}
	}
}
