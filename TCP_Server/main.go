package main

import (
	web "main/web"

	"github.com/sirupsen/logrus"
)

func main() {
	InitLogger()

	web.CreateTCPServer("8080", *Logger.WithFields(logrus.Fields{"prefix": "[TCP_SERVER]"}))
	web.StartWebServer("8000")
	web.StopTCPServer()
}
