package handlers

import (
	"github.com/assignments-zanewebbUW/servers/gateway/models/users"
	"github.com/assignments-zanewebbUW/servers/gateway/sessions"
)

//TODO: define a handler context struct that
//will be a receiver on any of your HTTP
//handler functions that need access to
//globals, such as the key used for signing
//and verifying SessionIDs, the session store
//and the user store

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
