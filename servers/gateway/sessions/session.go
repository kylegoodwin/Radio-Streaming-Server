package sessions

import (
	"errors"
	"log"
	"net/http"
	"strings"
)

const headerAuthorization = "Authorization"
const paramAuthorization = "auth"
const schemeBearer = "Bearer "

//ErrNoSessionID is used when no session ID was found in the Authorization header
var ErrNoSessionID = errors.New("no session ID found in " + headerAuthorization + " header")

//ErrInvalidScheme is used when the authorization scheme is not supported
var ErrInvalidScheme = errors.New("authorization scheme not supported")

//BeginSession creates a new SessionID, saves the `sessionState` to the store, adds an
//Authorization header to the response with the SessionID, and returns the new SessionID
func BeginSession(signingKey string, store Store, sessionState interface{}, w http.ResponseWriter) (SessionID, error) {
	//TODO:
	//- create a new SessionID
	sid, err := NewSessionID(signingKey)
	if err != nil {
		log.Println("Failed to create new sessionID for user")
		return InvalidSessionID, err
	}

	//- save the sessionState to the store
	err = store.Save(sid, sessionState)
	if err != nil {
		log.Println("Failed to save the SID to the store")
		return InvalidSessionID, err
	}

	//- add a header to the ResponseWriter that looks like this:
	//    "Authorization: Bearer <sessionID>"
	//  where "<sessionID>" is replaced with the newly-created SessionID
	//  (note the constants declared for you above, which will help you avoid typos)
	w.Header().Add(headerAuthorization, schemeBearer+sid.String())
	//log.Printf("Returned SID from BeginSession is: %v", sid)
	return sid, nil
	//return InvalidSessionID, nil
}

//GetSessionID extracts and validates the SessionID from the request headers
func GetSessionID(r *http.Request, signingKey string) (SessionID, error) {
	//TODO: get the value of the Authorization header,
	//or the "auth" query string parameter if no Authorization header is present,
	//and validate it. If it's valid, return the SessionID. If not
	//return the validation error.

	//Get the sid out of the request
	authHeader := r.Header.Get(headerAuthorization)
	if len(authHeader) == 0 {
		//if it was not found in the header check the query
		authHeader = r.URL.Query().Get(paramAuthorization)
	}

	//Check that scheme is correct
	if len(authHeader) != 0 && !strings.Contains(authHeader, schemeBearer) {
		log.Println("authorization header was not found")
		return InvalidSessionID, ErrInvalidScheme
	}

	//Check that there is an ID given
	sid := strings.Trim(authHeader, schemeBearer)
	if len(sid) == 0 {
		log.Println("SID was of length 0 from request")
		return InvalidSessionID, ErrInvalidID
	}

	//validate the sid
	validated, err := ValidateID(sid, signingKey)
	if err != nil {
		log.Println("SID did not validate")
		return InvalidSessionID, err
	}

	//if these are equal, the validation was successful
	if SessionID(sid) == validated {
		//log.Printf("Returned SID of %v to outer function", sid)
		return SessionID(sid), nil
	}
	return InvalidSessionID, nil
}

//GetState extracts the SessionID from the request,
//gets the associated state from the provided store into
//the `sessionState` parameter, and returns the SessionID
func GetState(r *http.Request, signingKey string, store Store, sessionState interface{}) (SessionID, error) {
	//TODO: get the SessionID from the request, and get the data
	//associated with that SessionID from the store.
	//log.Println("Directed to GetState")
	//Get SID
	sid, err := GetSessionID(r, signingKey)
	if err != nil {
		log.Printf("Didn't find the SID in the request. Error was: %v", err.Error())
		return InvalidSessionID, err
	}

	//Retrieve data from store
	err = store.Get(sid, sessionState)
	if err != nil {
		log.Printf("Didn't find the SID in the sessionStore. Error was: %v", err.Error())
		return InvalidSessionID, err
	}

	//Give SID
	return sid, nil
}

//EndSession extracts the SessionID from the request,
//and deletes the associated data in the provided store, returning
//the extracted SessionID.
func EndSession(r *http.Request, signingKey string, store Store) (SessionID, error) {
	//TODO: get the SessionID from the request, and delete the
	//data associated with it in the store.

	//Get SID
	sid, err := GetSessionID(r, signingKey)
	if err != nil {
		return InvalidSessionID, err
	}

	//Delete data from store
	err = store.Delete(sid)
	if err != nil {
		return InvalidSessionID, err
	}

	//Give SID
	return sid, nil
}
