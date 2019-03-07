package handlers

import (
	"github.com/Radio-Streaming-Server/servers/gateway/models/users"
	"github.com/Radio-Streaming-Server/servers/gateway/sessions"
)

//Context gives the RedisStore and MySQLStore so that they
//can be accessed globally
type Context struct {
	Key           string
	SessionsStore *sessions.RedisStore /**sessions.MemStore*/
	UsersStore    *users.MySQLStore    /* *users.MyMockStore*/
}

//NewContext creates a new context if given a key, sessionstore and userstore
func NewContext(key string, ss *sessions.RedisStore /* *sessions.MemStore*/, us *users.MySQLStore /* *users.MyMockStore*/) *Context {
	if ss == nil || us == nil || key == "" {
		return nil
	}

	context := Context{key, ss, us}
	return &context
}
