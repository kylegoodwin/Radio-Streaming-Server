package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync/atomic"

	"github.com/Radio-Streaming-Server/servers/gateway/sessions"
)

type Director func(r *http.Request)

func CustomDirector(targets []*url.URL, context HandlerContext) Director {
	var counter int32
	counter = 0
	return func(r *http.Request) {
		targ := targets[int(counter)%len(targets)]
		log.Println("here")
		log.Println(r)

		atomic.AddInt32(&counter, 1)
		r.Header.Add("X-Forwarded-Host", r.Host)
		//authHeader := r.Header.Get("Authorization")

		//Clear old header
		r.Header.Del("X-User")

		var sessionState *UserSession
		//Gets the users state from the sessionID, it gets populated into the sessionState
		_, err := sessions.GetState(r, context.Key, context.Session, &sessionState)

		//User properly authenticated, add x-user header
		if err == nil {
			user := sessionState.User
			log.Println("x-user-user")
			log.Println(user)
			marshaled, err := json.Marshal(user)

			if err != nil {
				fmt.Println("Error during json marshaling ")
			}
			r.Header.Set("X-User", string(marshaled))
		}
		r.Host = targ.Host
		r.URL.Host = targ.Host
		r.URL.Scheme = targ.Scheme

	}
}
