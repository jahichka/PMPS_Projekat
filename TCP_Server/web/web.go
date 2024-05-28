package web

import (
	"net/http"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const devHTML = `
<div class="ui raised segment">
  <tr id="%s">
    <td> %s (%s) </td>
    <td style="background-color: %s;"></td>
  </tr>
</div>
`

type Server struct {
	Engine      *gin.Engine
	Upgrader    websocket.Upgrader
	Connections map[string]*websocket.Conn
}

var webServer *Server

func StartWebServer(port string) {
	webServer = &Server{
		Engine: gin.Default(),
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		Connections: make(map[string]*websocket.Conn, 0),
	}

	webServer.Engine.LoadHTMLGlob("web/*.html")

	webServer.Engine.GET("/", webServer.homePage)
	webServer.Engine.GET("/ws", webServer.wsConn)

	webServer.Engine.Run(":" + port) // blocking function
}

func WSMessage(id, event, message, render string) error {
	for _, v := range webServer.Connections {
		if err := v.WriteJSON(map[string]any{"id": id, "event": event, "data": message, "render": render}); err != nil{
			return err
		}
	}
	return nil
}

func (webServer *Server) homePage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}

func (webServer *Server) wsConn(c *gin.Context) {
	conn, err := webServer.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	if _, exists := webServer.Connections[conn.RemoteAddr().String()]; exists {
		delete(webServer.Connections, conn.RemoteAddr().String())
	}
	webServer.Connections[conn.RemoteAddr().String()] = conn

	for _, dev := range tcpServer.devices {
		fmt.Println(dev)
		color := ""
		if ( dev.State == STATE_ON ) {
			color = COLOR_ON
		} else {
			color = COLOR_OFF
		}
		render := fmt.Sprintf(DEV_HTML, dev.Name, dev.ID, color)
		WSMessage(dev.ID, EVENT_STATE, "login", render )
	}
}
