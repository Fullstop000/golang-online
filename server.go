package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

func main() {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	upgrader := &websocket.Upgrader{
		HandshakeTimeout: 5 * time.Second,
		Subprotocols:     []string{"test_protocol"},
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	r.Use(func(context *gin.Context) {
		context.Set("upgrader", upgrader)
	})
	r.GET("/ws/go", wsHandler)
	r.Run(":8080")
}

func wsHandler(c *gin.Context) {
	u := c.MustGet("upgrader").(*websocket.Upgrader)
	conn, err := u.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Errorf("Error getting the websocket connection : %s", err)
		c.JSON(500, gin.H{"error": err})
		conn.Close()
		return
	}
	wb := &WebsocketBackend{
		connection: conn,
		recv:       make(chan *formattedLog),
		errCh:      make(chan error),
	}
	go wb.Write()
	data := wb.Read()
	bs, err := NewBuildSupport(data, wb)
	if err != nil {
		logger.Errorf("Error creating BuildSupport : %s", err)
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, fmt.Sprintf("Error creating BuildSupport : %s", err)))

		c.JSON(500, gin.H{"error": err})
		conn.Close()
		return
	}
	bs.Start()
}
