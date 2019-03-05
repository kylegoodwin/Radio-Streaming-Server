package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/assignments-zanewebbUW/servers/gateway/models/users"
	"github.com/assignments-zanewebbUW/servers/gateway/sessions"
)

//TODO: define HTTP handler functions as described in the
//assignment description. Remember to use your handler context
//struct as the receiver on these functions so that you have
//access to things like the session store and user store.

//UsersHandler : Generates a new user and session for that user (ONLY FOR POST REQUESTS)
func (cont *Context) UsersHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Directed to UsersHandler")
	if r.Method == "POST" {
		if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			//Grab JSON for new user from body and validate
			decoder := json.NewDecoder(r.Body)
			var nu users.NewUser

			err := decoder.Decode(&nu)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = nu.Validate()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			//turn new user into user
			u, err := nu.ToUser()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			//Add it to the user db
			uResp, err := cont.UsersStore.Insert(u)
			if err != nil {
				log.Printf("Insert Failed. Error was: %v", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if uResp.ID == 0 {
				log.Printf("ID not assigned. Error was: %v", err.Error())
				http.Error(w, "Database did not assign an ID to user", http.StatusInternalServerError)
				return
			}

			//Make sure that the user is findable in the Trie
			cont.UsersStore.InsertUserIntoTrie(uResp)

			//Generate and store new session
			newSessState := SessionState{time.Now(), *uResp}
			_, err = sessions.BeginSession(cont.Key, cont.SessionsStore, newSessState, w)
			if err != nil {
				log.Println("Failed to create session")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			//Generate successful response
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			encoder := json.NewEncoder(w)
			err = encoder.Encode(uResp)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		http.Error(w, "JSON is the only supported media type", http.StatusUnsupportedMediaType)
	}

	//Users search API
	if r.Method == "GET" {
		//Authentication check
		currentSessionState := &SessionState{} //Gotta make sure to ues the & thingy to make sure it's instantiated
		_, err := sessions.GetState(r, cont.Key, cont.SessionsStore, currentSessionState)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		//Grab the prefix and check that it isn't nothing
		queryPrefix := r.URL.Query().Get("q")
		if len(queryPrefix) == 0 {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		//Make sure to do this so the search actually works in all cases
		queryPrefix = strings.ToLower(queryPrefix)

		//Grab the users ID from the trie
		returnedUsers, err := cont.UsersStore.Trie.Find(20, queryPrefix)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//Add them to users object list
		var usersToSort []*users.User
		for _, id := range returnedUsers {
			//Grab the user from the DB
			user, err := cont.UsersStore.GetByID(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//Add it to the list
			usersToSort = append(usersToSort, user)
		}

		//Grab their names for sorting
		var userNames []string
		for _, u := range usersToSort {
			userNames = append(userNames, u.UserName)
		}

		//Sort them
		sort.Strings(userNames)

		//Populate a user slice with the correct order
		var sortedUsers []*users.User
		for _, un := range userNames {
			for _, u := range usersToSort {
				if u.UserName == un {
					sortedUsers = append(sortedUsers, u)
				}
			}
		}

		//Encode and return the objects
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(sortedUsers)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	http.Error(w, "GET and POST are the ony supported methods for this handler", http.StatusMethodNotAllowed)
}

//SpecificUserHandler either gets the information of a user, or updates the info of an authenticated user.
func (cont *Context) SpecificUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Directed to SpecificUsersHandler")
	//Authentication check
	currentSessionState := &SessionState{} //Gotta make sure to ues the & thingy to make sure it's instantiated
	_, err := sessions.GetState(r, cont.Key, cont.SessionsStore, currentSessionState)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method == "GET" {
		//Get the requested ID
		requestedID, err := FetchRequestedID(r, currentSessionState.User.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		//Look for the user
		returnedUser, err := cont.UsersStore.GetByID(requestedID)
		if err != nil {
			http.Error(w, "User was not found", http.StatusNotFound)
		}

		//Return the user in JSON
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		err = encoder.Encode(returnedUser)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	//Receives an Updates object and applies it to a stored user
	//This can only be performed on the authenticated user
	if r.Method == "PATCH" {
		if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			//Get the requested ID
			requestedID, err := FetchRequestedID(r, currentSessionState.User.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			//Begin handling the actual request after checking a few more things.
			if currentSessionState.User.ID == requestedID {

				//Decode the given updates from the request
				decoder := json.NewDecoder(r.Body)
				var givenUpdates users.Updates
				err := decoder.Decode(&givenUpdates)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				//Perform the update in the mySQL DB
				updatedUser, err := cont.UsersStore.Update(requestedID, &givenUpdates)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				//Fetch the user before the updates were performed so we can scrub the old first and last names
				preUpdateUser := currentSessionState.User
				cont.UsersStore.Trie.Remove(preUpdateUser.FirstName, preUpdateUser.ID)
				cont.UsersStore.Trie.Remove(preUpdateUser.LastName, preUpdateUser.ID)

				//Add the updated user. This will silently reject the username insert and add the first and last names
				cont.UsersStore.InsertUserIntoTrie(updatedUser)

				//Respond with the updated user
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				encoder := json.NewEncoder(w)
				encoder.Encode(updatedUser)

			}
			http.Error(w, "Cannot update a user that is not yourself", http.StatusForbidden)
		}
		http.Error(w, "JSON is the only supported media type", http.StatusUnsupportedMediaType)

	}
	http.Error(w, "GET and PATCH are the ony supported methods for this handler", http.StatusMethodNotAllowed)
}

//SessionsHandler Creates a new session for a user that provides the correct email and password of an existing user
func (cont *Context) SessionsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Directed to SessionsHandler")
	if r.Method == "POST" {
		if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			log.Println("Attempted Sign In")
			var givenCredentials users.Credentials
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&givenCredentials)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			//Try to find the user with the email given
			attemptingUser, err := cont.UsersStore.GetByEmail(givenCredentials.Email)
			if err != nil {
				time.Sleep(2 * time.Second)
				http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
				return
			}

			//Try to verify the password
			err = attemptingUser.Authenticate(givenCredentials.Password)
			if err != nil {
				http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
				return
			}

			//Create the new session
			newSessionState := SessionState{time.Now(), *attemptingUser}
			_, err = sessions.BeginSession(cont.Key, cont.SessionsStore, newSessionState, w)
			if err != nil {
				http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
				return
			}

			//Log Successful sign-in
			_, err = cont.UsersStore.InsertSignIn(attemptingUser.ID, r.RemoteAddr)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			//Generate the response
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			encoder := json.NewEncoder(w)
			err = encoder.Encode(attemptingUser)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		http.Error(w, "JSON is the only supported media type", http.StatusUnsupportedMediaType)
	}
	http.Error(w, "POST is the ony supported method for this handler", http.StatusMethodNotAllowed)
}

//SpecificSessionHandler Logs out the user when DELETE method is given
func (cont *Context) SpecificSessionHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Directed to SpecificSessionsHandler")
	//Authentication check
	currentSessionState := &SessionState{} //Gotta make sure to ues the & thingy to make sure it's instantiated
	_, err := sessions.GetState(r, cont.Key, cont.SessionsStore, currentSessionState)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method == "DELETE" {
		lastSegment := path.Base(r.URL.Path)
		if lastSegment == "mine" {
			_, err := sessions.EndSession(r, cont.Key, cont.SessionsStore)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			//Confimation message
			w.Write([]byte("signed out"))
		}
		http.Error(w, "Cannot end the session of another user", http.StatusForbidden)
	}
	http.Error(w, "DELETE is the ony supported method for this handler", http.StatusMethodNotAllowed)
}

//FetchRequestedID returns an int64 of the requested userID from the URL path
//to be used in other operations
func FetchRequestedID(r *http.Request, authenticatedUserID int64) (int64, error) {
	requestedID := path.Base(r.URL.Path)
	var finalReqestedID int64

	//Could be using me or an actual number
	if requestedID == "me" {
		//if it is "me" grab the ID of the currently authenticated user
		finalReqestedID = authenticatedUserID
	} else {
		finalReqestedIDInt, err := strconv.Atoi(requestedID)
		if err != nil {
			return 0, err
		}
		finalReqestedID = int64(finalReqestedIDInt)
	}
	return finalReqestedID, nil
}
