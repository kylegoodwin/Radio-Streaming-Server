package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/Radio-Streaming-Server/servers/gateway/sessions"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

//TODO: add a handler that upgrades clients to a WebSocket connection
//and adds that to a list of WebSockets to notify when events are
//read from the RabbitMQ server. Remember to synchronize changes
//to this list, as handlers are called concurrently from multiple
//goroutines.

//TODO: start a goroutine that connects to the RabbitMQ server,
//reads events off the queue, and broadcasts them to all of
//the existing WebSocket connections that should hear about
//that event. If you get an error writing to the WebSocket,
//just close it and remove it from the list
//(client went away without closing from
//their end). Also make sure you start a read pump that
//reads incoming control messages, as described in the
//Gorilla WebSocket API documentation:
//http://godoc.org/github.com/gorilla/websocket

type msg struct {
	Message string `json:"message"`
}

//SocketStore has a mutex and stores websocket connections as a map of userID:connection key value pairs
type SocketStore struct {
	Connections map[int64]*websocket.Conn
	Lock        *sync.Mutex
	Cont        *Context
}

type rabbitMsg struct {
	Type    string
	Content interface{}
	UserIDs *[]int64
}

//Robbing these
// Control messages for websocket
const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1

	// BinaryMessage denotes a binary data message.
	BinaryMessage = 2

	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10
)

// //NewSocketStore makes a new SocketStore
// func NewSocketStore() *SocketStore {
// 	var m map[int]*websocket.Conn
// 	m = make(map[int]*websocket.Conn)
// 	ss := &SocketStore{m, &sync.Mutex{}}
// 	return ss
// 	//return &SocketStore{}
// }

//InsertConnection inserts a new connection to the SocketStore with a given user ID and websocket connection
func (s *SocketStore) InsertConnection(userID int64, conn *websocket.Conn) int64 {
	s.Lock.Lock()
	//connID := len(s.Connections)
	// insert socket connection
	s.Connections[userID] = conn
	s.Lock.Unlock()
	return userID
}

//RemoveConnection removes a websocket connection from the context's SocketStore with a given connection
func (s *SocketStore) RemoveConnection(userID int64) {
	//var userToRemove int64
	s.Lock.Lock()
	// //Need to find key for map of the correct user
	// for k := range s.Connections {
	// 	if s.Connections[k] == conn {
	// 		userToRemove = k
	// 	}
	// }
	// insert socket connection
	delete(s.Connections, userID)
	s.Lock.Unlock()
}

//SendMessages broadcasts a message to all connections in the store
func (s *SocketStore) SendMessages(messages <-chan amqp.Delivery) {
	for {

		msg := <-messages
		givenMsg := &rabbitMsg{}
		err := json.Unmarshal(msg.Body, givenMsg)
		if err != nil {
			fmt.Printf("error during unmarshalling: %v", err)
		}

		// log.Printf("Type: %v", givenMsg.Type)
		// log.Printf("Content: %v", givenMsg.Content)
		// log.Printf("UserIDs: %v", givenMsg.UserIDs)

		s.Lock.Lock()
		if len(*givenMsg.UserIDs) == 0 {
			log.Print("givenMsg.UserIDs length was 0")
			for ID, c := range s.Connections {
				err := c.WriteMessage(TextMessage, msg.Body)
				if err != nil {
					s.RemoveConnection(ID)
				}
			}
		} else {
			for ID, c := range s.Connections {
				if contains(givenMsg.UserIDs, ID) {
					err := c.WriteMessage(TextMessage, msg.Body)
					if err != nil {
						s.RemoveConnection(ID)
					}
				}
			}
		}
		s.Lock.Unlock()

		// for _, conn := range s.Connections {
		// 	writeError = conn.WriteMessage(messageType, data)
		// 	if writeError != nil {
		// 		return writeError
		// 	}
		// }
	}
}

func contains(s *[]int64, e int64) bool {
	for _, a := range *s {
		if a == e {
			return true
		}
	}
	return false
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//SocketConnectionHandler will upgrade a connection to a websocket connection if the request has a valid session token
func (s *SocketStore) SocketConnectionHandler(w http.ResponseWriter, r *http.Request) {
	if false { //r.Header.Get("Origin") != "https://localhost" { //"https://api.zanewebb.me" { //I think this is supposed to be the client domain and not the api domain but
		http.Error(w, "Websocket Connection Refused", 403)

	} else {
		//Authentication check
		currentSessionState := &SessionState{} //Gotta make sure to ues the & thingy to make sure it's instantiated
		_, err := sessions.GetState(r, s.Cont.Key, s.Cont.SessionsStore, currentSessionState)
		if err != nil {
			log.Printf("Error during authentication before websocket upgrade: %v", err)
			http.Error(w, "Client is not logged in and cannot upgrade to websocket", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		s.InsertConnection(currentSessionState.User.ID, conn)
		//Kick off goroutine to check on the connection consistently
		go s.CheckAlive(conn, currentSessionState.User.ID)
	}
}

//CheckAlive makes sure that the websockets are still live and closes the connections if they are not.
func (s *SocketStore) CheckAlive(conn *websocket.Conn, userID int64) {
	defer conn.Close()
	defer s.RemoveConnection(userID)

	for {
		messageType, p, err := conn.ReadMessage()

		if messageType == TextMessage || messageType == BinaryMessage {
			fmt.Printf("Client says %v\n", p)
			// fmt.Printf("Writing %s to all sockets\n", string(p))
			// s.WriteToAllConnections(TextMessage, append([]byte("Hello from server: "), p...))
		} else if messageType == CloseMessage {
			fmt.Println("Close message received.")
			break
		} else if err != nil {
			fmt.Println("Error reading message.")
			break
		}
		// ignore ping and pong messages
	}

}
