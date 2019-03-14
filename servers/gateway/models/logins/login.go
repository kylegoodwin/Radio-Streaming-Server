package logins

import "time"

//Login represents a userlogin
type Login struct {
	Key       int64
	UserID    int64
	LoginTime time.Time
	UserIP    string
}
