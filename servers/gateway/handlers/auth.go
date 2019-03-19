package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Radio-Streaming-Server/servers/gateway/models/logins"

	"github.com/Radio-Streaming-Server/servers/gateway/sessions"

	"github.com/Radio-Streaming-Server/servers/gateway/models/users"
)

const HeaderContentType = "Content-Type"
const HeaderJSON = "application/json"
const HeaderAuth = "Authorization"

//TODO: define HTTP handler functions as described in the
//assignment description. Remember to use your handler context
//struct as the receiver on these functions so that you have
//access to things like the session store and user store.

//UsersHandler handles requests for the Users Resource, such as adding new users or getting a list of users
func (context *HandlerContext) UsersHandler(w http.ResponseWriter, r *http.Request) {
	//Handler requests for the users resource
	//Just post requests for now

	var session UserSession
	header := r.Method

	//Send request to search
	if header == http.MethodGet {
		context.UserQueryHandler(w, r)
		return
	}

	if header != http.MethodPost {
		http.Error(w, "Http Method"+header+"not allowed", http.StatusMethodNotAllowed)
		return
	}

	//Check that the body is JSON
	contentType := r.Header.Get(HeaderContentType)

	if contentType != HeaderJSON {
		http.Error(w, fmt.Sprintf("Request body must be in"+HeaderJSON+"format"), http.StatusUnsupportedMediaType)
		return
	}

	user := &users.NewUser{}
	err := json.NewDecoder(r.Body).Decode(user)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request body"), http.StatusBadRequest)
		return
	}

	err = user.Validate()

	if err != nil {
		http.Error(w, fmt.Sprintf("Error validating user"), http.StatusBadRequest)
		return
	}

	validatedUser, err := user.ToUser()

	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating user"), http.StatusInternalServerError)
		return
	}

	validatedUser, err = context.User.Insert(validatedUser)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error inserting user into database: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	//Append the new user to the trie
	names := []string{}
	firstNames := strings.Split(strings.ToLower(validatedUser.FirstName), " ")
	lastNames := strings.Split(strings.ToLower(validatedUser.LastName), " ")
	names = append(names, firstNames...)
	names = append(names, lastNames...)

	for _, name := range names {
		//It for some reason breaks without this print statement,
		//I have a feeling it is some weird concurrency / threading issue
		fmt.Println(name)
		context.Trie.Add(name, validatedUser.ID)
	}

	session.User = *validatedUser
	session.Usertime = time.Now()

	//This passes in context sessionStore
	_, err = sessions.BeginSession(context.Key, context.Session, session, w)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating a new session for the user"), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", HeaderJSON)
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(validatedUser)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error encoding the users JSON"), http.StatusInternalServerError)
		return
	}

}

//SpecificUserHandler handles the request for a request to a specific user
func (context *HandlerContext) SpecificUserHandler(w http.ResponseWriter, r *http.Request) {
	//Handler requests for the users resource
	//Just post requests for now

	var id int64
	idString := path.Base(r.URL.Path)

	var sessionState *UserSession

	//Gets the users state from the sessionID, it gets populated into the sessionState
	_, err := sessions.GetState(r, context.Key, context.Session, &sessionState)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting users session state: %s", err.Error()), http.StatusUnauthorized)
		return
	}

	if idString == "me" {
		id = sessionState.User.ID
	} else {
		convID, err := strconv.Atoi(idString)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to convert UserID from a string to an int"),
				http.StatusBadRequest)
			return
		}
		id = int64(convID)
	}

	method := r.Method

	//If the request is a GET

	if method == http.MethodGet {
		//Get the user profile associated with the requested user ID from your store.
		user, err := context.User.GetByID(id)

		if err != nil || user == nil {
			//If no user is found with that ID, respond with an http.StatusNotFound (404) status code, and a suitable message.
			http.Error(w, fmt.Sprintf("No user account assoicated with that ID"), http.StatusNotFound)
			return
		}

		//Otherwise, respond to the client with:
		//a status code of http.StatusOK (200).
		//a response Content-Type header set to application/json to indicate that the response body is encoded as JSON
		//the users.User struct returned by your store in the response body, encoded as a JSON object.

		w.Header().Set(HeaderContentType, HeaderJSON)
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(user)

		if err != nil {
			http.Error(w, fmt.Sprintf("Error encoding the users JSON"), http.StatusInternalServerError)
			return
		}

	} else if method == http.MethodPatch {

		//This should work because you get the id from "me"s sessionState earlier
		if sessionState.User.ID != id {
			http.Error(w, fmt.Sprintf("Access Denied. You must be authenticated as the user you are trying to update"), http.StatusForbidden)
			return

		}

		//If the request's Content-Type header does not start with application/json, respond with status code
		//http.StatusUnsupportedMediaType (415), and a message indicating that the request body must be in JSON.
		if r.Header.Get(HeaderContentType) != HeaderJSON {
			http.Error(w, fmt.Sprintf("Request body must be in JSON, instead received %s", r.Header.Get(HeaderContentType)), http.StatusForbidden)
			return
		}

		//Get user for deleting from the trie
		oldUser, _ := context.User.GetByID(id)

		//The request body should contain JSON that can be decoded into the users.Updates struct.
		var update users.Updates
		err := json.NewDecoder(r.Body).Decode(&update)

		if err != nil {
			http.Error(w, fmt.Sprintf("Error decoding request body"), http.StatusBadRequest)
			return
		}

		//Use that to update the user's profile.

		err = sessionState.User.ApplyUpdates(&update)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid Update: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		user, err := context.User.Update(id, &update)

		if err != nil {
			http.Error(w, fmt.Sprintf("Error updating user in database: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		//Remove old values from the trie
		oldNames := []string{}
		oldFirstNames := strings.Split(strings.ToLower(oldUser.FirstName), " ")
		oldLastNames := strings.Split(strings.ToLower(oldUser.LastName), " ")
		oldNames = append(oldNames, oldFirstNames...)
		oldNames = append(oldNames, oldLastNames...)

		for _, name := range oldNames {
			context.Trie.Remove(name, oldUser.ID)
		}

		//Append the new user to the trie
		newNames := []string{}
		newFirstNames := strings.Split(strings.ToLower(user.FirstName), " ")
		newLastNames := strings.Split(strings.ToLower(user.LastName), " ")
		newNames = append(newNames, newFirstNames...)
		newNames = append(newNames, newLastNames...)

		for _, name := range newNames {
			context.Trie.Add(name, user.ID)
		}

		//a status code of http.StatusOK (200).

		//a response Content-Type header set to application/json to indicate that the response body is encoded as JSON.
		w.Header().Set(HeaderContentType, HeaderJSON)
		w.WriteHeader(http.StatusOK)

		//a full copy of the updated user profile in the response body, encoded as a JSON object.
		err = json.NewEncoder(w).Encode(user)

		if err != nil {
			http.Error(w, fmt.Sprintf("Error encoding the users JSON"), http.StatusInternalServerError)
			return
		}

	} else {
		http.Error(w, fmt.Sprintf("HttpMethod %s Not Supported", method), http.StatusMethodNotAllowed)
		return
	}

}

//SessionsHandler handles requests for the "sessions" resource, and allows clients to begin a
//new session using an existing user's credentials.
//Allows clients to begin a new session using an existing user's credentials.
func (context *HandlerContext) SessionsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Http Method"+r.Method+"not allowed", http.StatusMethodNotAllowed)
		return
	}

	//Check that the body is JSON
	contentType := r.Header.Get(HeaderContentType)

	if contentType != HeaderJSON {
		http.Error(w, fmt.Sprintf("Request body must be in %s format instead was %s", HeaderJSON, contentType), http.StatusUnsupportedMediaType)
		return
	}

	var credentials users.Credentials
	err := json.NewDecoder(r.Body).Decode(&credentials)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request body"), http.StatusBadRequest)
		return
	}

	user, err := context.User.GetByEmail(credentials.Email)

	if err != nil {
		http.Error(w, fmt.Sprintf("invalid credentials"), http.StatusUnauthorized)
		return
	}

	err = user.Authenticate(credentials.Password)

	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, fmt.Sprintf("invalid credentials"), http.StatusUnauthorized)
		return
	}

	//Inserting into database

	login := logins.Login{}
	login.LoginTime = time.Now()
	login.UserID = user.ID

	xForward := r.Header.Get("X-Forwarded-For")
	if xForward != "" {
		login.UserIP = strings.Split(xForward, ",")[0]
	} else {
		login.UserIP = r.RemoteAddr
	}

	context.Login.Insert(&login)

	//Begin a new session
	var sessionState UserSession
	sessionState.User = *user
	sessionState.Usertime = time.Now()

	sessions.BeginSession(context.Key, context.Session, sessionState, w)

	w.Header().Set(HeaderContentType, HeaderJSON)
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(user)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error encoding the users JSON"), http.StatusInternalServerError)
		return
	}

}

//SpecificSessionHandler
//This function handles requests related to a specific authenticated session.
func (context *HandlerContext) SpecificSessionHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodDelete {
		url := path.Base(r.URL.Path)

		if url != "mine" {
			http.Error(w, fmt.Sprintf("You cannot access sessions that are not yours"), http.StatusForbidden)
			return
		}

		_, err := sessions.EndSession(r, context.Key, context.Session)

		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected Error Ending Session"), http.StatusInternalServerError)
			return
		}

		w.Header().Add(HeaderContentType, "text/plain")
		io.WriteString(w, "signed out")
	} else {
		http.Error(w, fmt.Sprintf("Method not allowed"), http.StatusMethodNotAllowed)

	}
}

//UserQueryHandler handles the Trie search from an authenticated user
func (context *HandlerContext) UserQueryHandler(w http.ResponseWriter, r *http.Request) {

	var sessionState *UserSession

	//Gets the users state from the sessionID, it gets populated into the sessionState
	_, err := sessions.GetState(r, context.Key, context.Session, &sessionState)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting users session state: %s", err.Error()), http.StatusUnauthorized)
		return
	}

	//Get all users query
	if len(r.URL.Query()["all"]) != 0 {
		if r.URL.Query()["all"][0] == "true" {
			allUsers, err := context.User.GetAllUsers()

			w.Header().Set(HeaderContentType, HeaderJSON)
			err = json.NewEncoder(w).Encode(allUsers)

			if err != nil {
				http.Error(w, fmt.Sprintf("Error encoding the users JSON"), http.StatusInternalServerError)
				return
			}
			return

		}
	}

	if len(r.URL.Query()["q"]) == 0 {
		http.Error(w, "No query given", http.StatusBadRequest)
		return
	}

	query := r.URL.Query()["q"][0]

	query = strings.ToLower(query)

	values := context.Trie.Find(query, 20)

	users := []*users.User{}

	for _, user := range values {
		user, err := context.User.GetByID(user)
		if err != nil {
			http.Error(w, "Trie returned invalid user id", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].UserName < users[j].UserName
	})

	w.Header().Set(HeaderContentType, HeaderJSON)
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(users)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error encoding the users JSON"), http.StatusInternalServerError)
		return
	}
}
