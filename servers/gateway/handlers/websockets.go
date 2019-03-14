package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/streadway/amqp"

	"github.com/gorilla/websocket"
	"github.com/Radio-Streaming-Server/servers/gateway/sessions"
)

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

func (context *HandlerContext) WSUpgrade(w http.ResponseWriter, r *http.Request) {
	var sessionState *UserSession
	//Gets the users state from the sessionID, it gets populated into the sessionState
	_, err := sessions.GetState(r, context.Key, context.Session, &sessionState)

	//User properly authenticated, add x-user header

	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	userID := sessionState.User.ID

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// I dont have a client yet, so this doesnt matter
			/*
				fmt.Println("origin check")
				fmt.Println(r.Header.Get("Origin"))
				if r.Header.Get("Origin") != "https://api.kylegoodwin.net" {
					return false
				}
			*/

			return true
		},
	}

	if !upgrader.CheckOrigin(r) {
		http.Error(w, "Websocket Connection Refused", 403)
	} else {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Failed to open websocket connection", 401)
			return
		}
		context.Sockets.InsertConnection(conn, userID)

		connId := userID
		// Invoke a goroutine for handling control messages from this connection
		go (func(conn *websocket.Conn, connId int64) {
			defer conn.Close()
			defer context.Sockets.RemoveConnection(connId)

			for {
				messageType, p, err := conn.ReadMessage()

				if messageType == TextMessage || messageType == BinaryMessage {
					fmt.Printf("Client says %v\n", p)
				} else if messageType == CloseMessage {
					fmt.Println("Close message received.")
					conn.Close()
					break
				} else if err != nil {
					fmt.Println(err)
					fmt.Println("Error reading message.")
					break
				}
				// ignore ping and pong messages
			}

		})(conn, connId)

	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type message struct {
	UserIDs []int64 `json:"userIDs"`
}

func (sockets *userSockets) SendMessages(messages <-chan amqp.Delivery) {
	for msg := range messages {

		//result := message{}
		var result message
		err := json.Unmarshal(msg.Body, &result)

		if err != nil {
			fmt.Println(err)
		}

		if len(result.UserIDs) == 0 {
			for _, conn := range sockets.Notifier {
				conn.WriteMessage(TextMessage, msg.Body)
			}
		} else {
			for _, user := range result.UserIDs {

				if currentSocket, ok := sockets.Notifier[user]; ok {
					fmt.Println("Wrote message to user : " + string(user))
					err := currentSocket.WriteMessage(TextMessage, msg.Body)

					if err != nil {
						fmt.Println(err)
					}

				} else {
					fmt.Println("no user on socket " + string(user))

				}

				if err != nil {
					fmt.Println(err)
				}
			}

		}

		test := result.UserIDs
		fmt.Println(test)

	}

}

// Thread-safe method for inserting a connection
func (s *userSockets) InsertConnection(conn *websocket.Conn, user int64) {
	s.Lock.Lock()

	// insert socket connection
	s.Notifier[user] = conn
	s.Lock.Unlock()

}

// Thread-safe method for inserting a connection
func (s *userSockets) RemoveConnection(connId int64) {
	s.Lock.Lock()
	// insert socket connection
	delete(s.Notifier, connId)
	s.Lock.Unlock()
}

func EstablishSockets() *userSockets {
	conns := make(map[int64]*websocket.Conn)
	conn := userSockets{}
	conn.Notifier = conns
	return &conn
}
