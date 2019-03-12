package handlers

import (
	"sync"

	"github.com/gorilla/websocket"
)

type userSockets struct {
	Notifier map[int64]*websocket.Conn
	Lock     sync.Mutex
}
