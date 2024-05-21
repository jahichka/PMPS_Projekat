package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

type CustomFormatter struct {
	Caller string
}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	message := fmt.Sprintf("[%s][%s] TCP_SERVER : %s\n",
		entry.Time.Format("2006-01-02 15:04:05"),
		entry.Level.String(),
		entry.Message,
	)

	return []byte(message), nil
}

func InitLogger() {
	Logger = logrus.New()
	Logger.SetFormatter(&CustomFormatter{})
	Logger.SetOutput(os.Stdout)
}
