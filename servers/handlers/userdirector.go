package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"sync/atomic"

	"github.com/assignments-zanewebbUW/servers/gateway/sessions"
)

//Director wkjdnakjwdnkjawd
type Director func(r *http.Request)

//UserDirector setup needed for X-User header
func (cont *Context) UserDirector(targets []*url.URL) Director {
	var counter int32
	counter = 0

	return func(r *http.Request) {
		//increments only up to the length of the given array thanks to modulo
		targ := targets[counter%int32(len(targets))]
		atomic.AddInt32(&counter, 1)

		r.Header.Del("X-User")
		log.Println("Deleted existing X-User header")

		//Authentication check
		currentSessionState := &SessionState{} //Gotta make sure to ues the & thingy to make sure it's instantiated
		_, err := sessions.GetState(r, cont.Key, cont.SessionsStore, currentSessionState)
		if err != nil {
			log.Printf("Error during GetState: %v", err)
			return
		}

		newXUser := currentSessionState.User
		//Turn that object into a string for the header
		marshaled, err := json.Marshal(newXUser)
		if err != nil {
			log.Printf("Error during user marshalling: %v", err)
			return
		}

		log.Printf("Marshalled user: %s", string(marshaled))

		r.Host = targ.Host
		r.URL.Host = targ.Host
		r.URL.Scheme = targ.Scheme

		//Attach the json string to the header
		r.Header.Add("X-User", string(marshaled))

	}
}
