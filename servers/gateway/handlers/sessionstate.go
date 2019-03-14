package handlers

import (
	"time"

	"github.com/Radio-Streaming-Server/servers/gateway/models/users"
)

//TODO: define a session state struct for this web server
//see the assignment description for the fields you should include
//remember that other packages can only see exported fields!

//UserSession Defines the authentication session of a user
type UserSession struct {
	Usertime time.Time  `json:"time,omitempty"`
	User     users.User `json:"user,omitempty"`
}
