package main

import (
	"github.com/gorilla/websocket"
	"github.com/op/go-logging"
	"sync"
)

type formattedLog struct {
	Record  *logging.Record
	Message string
}
type WebsocketBackend struct {
	writingMux sync.Mutex
	connection *websocket.Conn
	recv       chan *formattedLog
	errCh      chan error
}

func (wb *WebsocketBackend) Log(level logging.Level, calldepth int, rec *logging.Record) error {
	wb.recv <- &formattedLog{
		Record:  rec,
		Message: rec.Formatted(calldepth + 1),
	}
	return nil
}
func (wb *WebsocketBackend) Close() {
	//close(wb.recv)
	//close(wb.errCh)
	wb.writingMux.Lock()
	err := wb.connection.WriteMessage(websocket.CloseMessage, []byte("  Finish "))
	wb.writingMux.Unlock()
	if err != nil {
		logger.Errorf("Error writing close message to the websocket : %s", err)
	}
	if err = wb.connection.Close(); err != nil {
		logger.Errorf("Error closing the websocket connection : %s", err)
	}
}
func (wb *WebsocketBackend) AbnormalClose(e error) {
	logger.Warningf("Abnormal closing action caused by err : %s", e)
	//close(wb.recv)
	//close(wb.errCh)
	wb.writingMux.Lock()
	if err := wb.connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, e.Error())); err != nil {
		logger.Errorf("Error writing CloseMessage to the websocket : %s", err)
	}
	wb.writingMux.Unlock()
	if err := wb.connection.Close(); err != nil {
		logger.Errorf("Error closing the websocket connection : %s", err)
	}
}
func (wb *WebsocketBackend) Read() []byte {
	_, data, err := wb.connection.ReadMessage()
	if err != nil {
		logger.Errorf("Error receiving the message from ws connection : %s", err)
		wb.errCh <- err
		return nil
	} else {
		return data
	}
}
func (wb *WebsocketBackend) Write() {
	for {
		select {
		case msg := <-wb.recv:
			//logger.Debugf("Receiving the log message : %s", msg)
			if msg.Record.Level == logging.WARNING && msg.Record.Args[0] == "close" {
				wb.Close()
			} else {
				wb.writingMux.Lock()
				if err := wb.connection.WriteMessage(websocket.TextMessage, []byte(msg.Message)); err != nil {
					logger.Errorf("Error writing message to the ws connection : %s", err)
					wb.errCh <- err
				}
				wb.writingMux.Unlock()
			}
		case err := <-wb.errCh:
			wb.AbnormalClose(err)
			break
		}
	}
}
