package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	InitLogger()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	ctx, stopServer := context.WithCancel(context.Background())

	server, err := NewTCPServer(ctx, ":8080", *Logger.WithFields(logrus.Fields{"prefix": "[TCP SERVER]"}))
	if err != nil {
		Logger.Fatalln("Server couldn't be created, exiting with error: ", err.Error())
	}

	server.Start()

	<-stop // block further execution until kill signal is passed
	stopServer()
	Logger.Info(" Shutdown initiated")
	server.Wg.Wait()
	Logger.Info("\n-- -- Shutdown complete -- --\n")

}

func parseArgsServer() (*string, error) {
	if len(os.Args) < 2 {
		return nil, errors.New("port not provided")
	}
	port := ":" + os.Args[1]
	return &port, nil
}
