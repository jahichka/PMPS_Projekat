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

	DEV_HTML_FULL = `
		<tr id=%s>
    	<td> %s (%s) </td>
    	<td> <button class="ui basic button" style="border: 0px;" onclick='openOverlay("%[1]s")'> Set Parameters </i> </td>
    	<td style="background-color: %[4]s;"></td>
		<tr id=%s>
	`
	// <div class="ui basic icon buttons" type="submit" onclick='createMatrix("%[1]s")'>
	// 	<button class="ui button" style="border: none;"><i class="upload icon"></i></button>
	// </button>

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
	DevID  string
	Buffer []byte
	Size   int
}

type RecievedData struct {
	AngleCount    int     `json:"angle_count"`
	BladeAngles   []int   `json:"blade_angles"`
	ControlValues [][]int `json:"control_values"`
	DevID         string  `json:"dev_id"`
	WindCount     int     `json:"wind_count"`
	WindSpeeds    []int   `json:"wind_speeds"`
}

func (srw *TCPServer) SendEvent(devId string, message string) {
	currentTime := time.Now()
	render := fmt.Sprintf(EVENT_HTML, currentTime.Format("2006-01-02 3:4:5"), devId, message)
	WSMessage(devId, EVENT_MSG, message, render)
}

func (srw *TCPServer) SendMessage(sender, message string) {
	srw.logger.Infof("From %s : %s", sender, message)
	render := fmt.Sprintf(EVENT_HTML, time.Now().Format("2006-01-02 3:4:5"), sender, " "+message)
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
				srw.SendEvent(msg.DevID, strings.TrimSuffix(string(msg.Buffer[:msg.Size]), "\n"))
			}
		}
	}()

	srw.logger.Info("Message channel created !")
	return ch
}
