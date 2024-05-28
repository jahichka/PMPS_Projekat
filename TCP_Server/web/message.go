package web

import (
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	EVENT_STATE = "state"
	EVENT_MSG   = "message"
	EVENT_LOGIN = "login"

	STATE_OFF = 0
	STATE_ON  = 1

	COLOR_ON  = "#77dd77"
	COLOR_OFF = "#ff6961"

	DEV_HTML = `
    <td> %s (%s) </td>
    <td style="background-color: %s;"></td>
	`

	DEV_HTML_FULL = "<tr id=%s>" + DEV_HTML + "</tr>"

	EVENT_HTML = `
    	<div class="content">
      	<div class="date">
        	%s
      	</div>
      	<div class="summary">
         	 <a>%s</a> %s
      	</div>
    	</div>
	`
)

type Message struct {
	Conn   net.Conn
	Dev    *Device
	Buffer []byte
	Size   int
}

func (srw *TCPServer) SendEvent(dev *Device, message string) {
	currentTime := time.Now()
	render := fmt.Sprintf(EVENT_HTML, currentTime.Format("2006-01-02 3:4:5"), dev.ID, message)
	WSMessage(dev.ID, EVENT_MSG, message, render)
}

func (srw *TCPServer) SendMessage( sender, message string ){
	srw.logger.Infof("From %s : %s", sender, message)
	render := fmt.Sprintf(EVENT_HTML, time.Now().Format("2006-01-02 3:4:5"), sender, " " + message )
	WSMessage(message, EVENT_MSG, fmt.Sprintf("from %s", sender), render)
}

func (srw *TCPServer) messageSender() chan *Message {
	ch := make(chan *Message, 32)
	go func() {
		for {
			select {
			case msg := <-ch:
				logMsg := fmt.Sprintf("Received message from %s - '%s'", msg.Conn.RemoteAddr().String(), strings.TrimSuffix(string(msg.Buffer[:msg.Size]), "\n"))
				srw.logger.Info(logMsg)
				srw.SendEvent(msg.Dev, strings.TrimSuffix(string(msg.Buffer[:msg.Size]), "\n"))
			}
		}
	}()

	srw.logger.Info("Message channel created !")
	return ch
}
