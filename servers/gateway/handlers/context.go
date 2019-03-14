package handlers

import (
	"github.com/Radio-Streaming-Server/servers/gateway/indexes"
	"github.com/Radio-Streaming-Server/servers/gateway/models/logins"
	"github.com/Radio-Streaming-Server/servers/gateway/models/users"
	"github.com/Radio-Streaming-Server/servers/gateway/sessions"
)

//TODO: define a handler context struct that
//will be a receiver on any of your HTTP
//handler functions that need access to
//globals, such as the key used for signing
//and verifying SessionIDs, the session store
//and the user store

//Handler context stores the session and user stores so that http writers can access them
type HandlerContext struct {
	Key     string
	Session sessions.Store
	User    users.Store
	Login   logins.Store
	Trie    *indexes.Trie
	Sockets *userSockets
}
