package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Radio-Streaming-Server/servers/gateway/models/logins"

	"github.com/Radio-Streaming-Server/servers/gateway/sessions"

	"github.com/Radio-Streaming-Server/servers/gateway/models/users"
)

/*
 Remember to test not only correct inputs, but also incorrect inputs.
 Check not only the response body, but also the response status code and headers.
 Ensure that your handlers do the right thing in all cases.
*/

var defaultSessionDuration = time.Duration(time.Hour)

func TestUsersHandler(t *testing.T) {

	fakeConn := users.NewFakeConnection()
	fakeSession := sessions.NewMemStore(defaultSessionDuration, defaultSessionDuration)

	validNewUser := users.NewUser{
		Email:        "coolguy@ding.dong",
		Password:     "password",
		PasswordConf: "password",
		UserName:     "coolguy420",
		FirstName:    "Cool",
		LastName:     "Guy",
	}

	invalidNewUser := users.NewUser{
		Email:    "",
		Password: "fart",
		LastName: "Guy",
	}

	validUser, _ := validNewUser.ToUser()

	var handleTest = HandlerContext{
		Key:     "fakesigningkey",
		User:    fakeConn,
		Session: fakeSession,
	}

	cases := []struct {
		name                string
		method              string
		requestBody         users.NewUser
		expectedStatusCode  int
		expectedError       bool
		expectedContentType string
		expectedReturn      *users.User
	}{

		{
			"Valid Post Request",
			http.MethodPost,
			validNewUser,
			http.StatusCreated,
			false,
			HeaderJSON,
			validUser,
		},

		{
			"Invalid User Sent",
			http.MethodPost,
			invalidNewUser,
			http.StatusBadRequest,
			true,
			HeaderJSON,
			validUser,
		},

		{
			"Invalid request body type (non json)",
			http.MethodPost,
			invalidNewUser,
			http.StatusUnsupportedMediaType,
			true,
			"text/plain; charset=utf-8",
			nil,
		},

		{
			"Invalid http method",
			http.MethodGet,
			validNewUser,
			http.StatusMethodNotAllowed,
			true,
			"text/plain; charset=utf-8",
			nil,
		},
	}

	for _, c := range cases {
		body, _ := json.Marshal(c.requestBody)
		request := httptest.NewRequest(c.method, "/v1/users", strings.NewReader(string(body)))
		request.Header.Set(HeaderContentType, c.expectedContentType)
		recorder := httptest.NewRecorder()

		handleTest.UsersHandler(recorder, request)
		response := recorder.Result()

		//Test Content type
		resContentType := response.Header.Get(HeaderContentType)
		if !c.expectedError && c.expectedContentType != resContentType {
			t.Errorf("case %s: incorrect return type: expected %s but recieved %s", c.name, c.expectedContentType, resContentType)

		}

		resStatusCode := response.StatusCode
		if c.expectedStatusCode != resStatusCode {
			t.Errorf("case %s: incorrect status code: expected: %d recieved: %d",
				c.name, c.expectedStatusCode, resStatusCode)
		}

		user := &users.User{}
		err := json.NewDecoder(response.Body).Decode(user)
		if c.expectedError && err == nil {
			t.Errorf("case %s: expected error but revieved none", c.name)
		}

		if !c.expectedError && c.expectedReturn.Email != user.Email && string(c.expectedReturn.PassHash) != string(user.PassHash) &&
			c.expectedReturn.FirstName != user.FirstName && c.expectedReturn.LastName != user.LastName {
			t.Errorf("case %s: incorrect return: expected %v but revieved %v",
				c.name, c.expectedReturn, user)
		}

	}

}

//Need to fix this, problems authenticating the user
func TestSpecificUserHandler(t *testing.T) {
	fakeConn := users.NewFakeConnection()
	fakeSession := sessions.NewMemStore(defaultSessionDuration, defaultSessionDuration)

	var handleTest = HandlerContext{
		Key:     "fakesigningkey",
		User:    fakeConn,
		Session: fakeSession,
	}

	validNewUser := users.NewUser{
		Email:        "coolguy@gmail.com",
		Password:     "password",
		PasswordConf: "password",
		UserName:     "coolguy",
		FirstName:    "Cool",
		LastName:     "Guy",
	}

	validUser, _ := validNewUser.ToUser()

	validUpdates := users.Updates{
		FirstName: "New",
		LastName:  "Name",
	}

	invalidUpdates := users.Updates{
		FirstName: "",
		LastName:  "",
	}

	updatedUser := users.User{
		ID:        1,
		Email:     "coolguy@gmail.com",
		FirstName: "New",
		LastName:  "Name",
	}
	updatedUser.SetPassword("password")

	cases := []struct {
		name                string
		method              string
		requestUserID       string
		requestBody         users.Updates
		expectedStatusCode  int
		expectedError       bool
		expectedContentType string
		expectedReturn      *users.User
	}{

		{
			"User is authenticated, Get ",
			http.MethodGet,
			"1",
			validUpdates,
			http.StatusOK,
			false,
			HeaderJSON,
			validUser,
		},
		{
			"User is authenticated, Get me ",
			http.MethodGet,
			"me",
			validUpdates,
			http.StatusOK,
			false,
			HeaderJSON,
			validUser,
		},
		{
			"Bad method",
			http.MethodDelete,
			"1",
			validUpdates,
			http.StatusMethodNotAllowed,
			true,
			HeaderJSON,
			validUser,
		},
		{
			"User is authenticated, Patch",
			http.MethodPatch,
			"0",
			validUpdates,
			http.StatusOK,
			false,
			HeaderJSON,
			&updatedUser,
		},
		{
			"Trying to update other user, Patch",
			http.MethodPatch,
			"1",
			validUpdates,
			http.StatusForbidden,
			true,
			HeaderJSON,
			&updatedUser,
		},
		{
			"Bad update, Patch",
			http.MethodPatch,
			"0",
			invalidUpdates,
			http.StatusInternalServerError,
			true,
			HeaderJSON,
			&updatedUser,
		},
	}

	for _, c := range cases {

		reqBody, _ := json.Marshal(c.requestBody)
		var request *http.Request

		if c.method == http.MethodPatch {
			request = httptest.NewRequest(c.method, "/v1/users/"+string(c.requestUserID), strings.NewReader(string(reqBody)))
			request.Header.Set(HeaderContentType, HeaderJSON)

		} else {
			request = httptest.NewRequest(c.method, "/v1/users/"+string(c.requestUserID), nil)
		}
		recorder := httptest.NewRecorder()

		var sessionState UserSession
		sessionState.User = *validUser
		sessionState.Usertime = time.Now()

		sid, _ := sessions.BeginSession(handleTest.Key, handleTest.Session, &sessionState, recorder)

		request.Header.Set(HeaderAuth, "Bearer "+string(sid))
		handleTest.SpecificUserHandler(recorder, request)

		response := recorder.Result()
		/*
			fmt.Println(c.name)
			fmt.Println("--------------------")
			buf := new(bytes.Buffer)
			buf.ReadFrom(response.Body)
			newStr := buf.String()

			fmt.Printf(newStr)
			fmt.Println("--------------------")
		*/

		//Test Content type
		fmt.Println("&&&& " + c.name)
		fmt.Println(response.Header)
		resContentType := response.Header.Get(HeaderContentType)
		resStatusCode := response.StatusCode

		if !c.expectedError && c.expectedContentType != resContentType {
			t.Errorf("case %s: incorrect return type: expected %s but recieved %s", c.name, c.expectedContentType, resContentType)

		}

		if c.expectedStatusCode != resStatusCode {
			t.Errorf("case %s: incorrect status code: expected: %d recieved: %d",
				c.name, c.expectedStatusCode, resStatusCode)
		}

		user := &users.User{}
		err := json.NewDecoder(response.Body).Decode(user)

		if c.expectedError && err == nil {
			t.Errorf("case %s: expected error but revieved none", c.name)
		}

		if !c.expectedError && c.expectedReturn.Email != user.Email && string(c.expectedReturn.PassHash) != string(user.PassHash) &&
			c.expectedReturn.FirstName != user.FirstName && c.expectedReturn.LastName != user.LastName {
			t.Errorf("case %s: incorrect return: expected %v but revieved %v",
				c.name, c.expectedReturn, user)
		}

	}

}

func TestSessionsHandler(t *testing.T) {

	fakeConn := users.NewFakeConnection()
	fakeSession := sessions.NewMemStore(defaultSessionDuration, defaultSessionDuration)
	loginStore := logins.NewFakeConnection()

	var handleTest = HandlerContext{
		Key:     "fakesigningkey",
		User:    fakeConn,
		Session: fakeSession,
		Login:   loginStore,
	}

	validNewUser := users.NewUser{
		Email:        "coolguy@gmail.com",
		Password:     "password",
		PasswordConf: "password",
		UserName:     "coolguy",
		FirstName:    "Cool",
		LastName:     "Guy",
	}

	validUser, _ := validNewUser.ToUser()

	validUserCredentials := users.Credentials{
		Email:    "coolguy@gmail.com",
		Password: "password",
	}

	invalidUserCredentials := users.Credentials{
		Email:    "cool@gmail.com",
		Password: "notright",
	}

	cases := []struct {
		name                string
		method              string
		requestContentType  string
		requestBody         users.Credentials
		validUser           *users.User
		expectedStatusCode  int
		expectedError       bool
		expectedContentType string
		expectedReturn      *users.User
	}{

		{
			"Valid Post Request",
			http.MethodPost,
			HeaderJSON,
			validUserCredentials,
			validUser,
			http.StatusCreated,
			false,
			HeaderJSON,
			validUser,
		},

		{
			"Get Request (Shouldnt work)",
			http.MethodGet,
			HeaderJSON,
			validUserCredentials,
			validUser,
			http.StatusMethodNotAllowed,
			true,
			"text/plain",
			nil,
		},

		{
			"Non JSON Header",
			http.MethodPost,
			"text/plain",
			validUserCredentials,
			validUser,
			http.StatusUnsupportedMediaType,
			true,
			"text/plain",
			nil,
		},

		{
			"Invalid Usercredentials",
			http.MethodPost,
			HeaderJSON,
			invalidUserCredentials,
			validUser,
			http.StatusUnauthorized,
			true,
			"text/plain",
			nil,
		},
	}

	for _, c := range cases {
		//We dont actually need to insert because of mock database handleTest.User.Insert()

		//Make a request with the validuser credentials
		body, _ := json.Marshal(c.requestBody)
		request := httptest.NewRequest(c.method, "/v1/sessions", strings.NewReader(string(body)))
		request.Header.Set(HeaderContentType, c.requestContentType)

		recorder := httptest.NewRecorder()

		handleTest.SessionsHandler(recorder, request)
		response := recorder.Result()

		//Test Content type
		resContentType := response.Header.Get(HeaderContentType)
		if !c.expectedError && c.expectedContentType != resContentType {
			t.Errorf("case %s: incorrect return type: expected %s but recieved %s", c.name, c.expectedContentType, resContentType)
		}

		resStatusCode := response.StatusCode
		if c.expectedStatusCode != resStatusCode {
			t.Errorf("case %s: incorrect status code: expected: %d recieved: %d",
				c.name, c.expectedStatusCode, resStatusCode)
		}

		user := &users.User{}
		err := json.NewDecoder(response.Body).Decode(user)
		if c.expectedError && err == nil {
			t.Errorf("case %s: expected error but revieved none", c.name)
		}

		//Make sure you get back whats in the fake db
		if !c.expectedError && c.expectedReturn.Email != user.Email && string(c.expectedReturn.PassHash) != string(user.PassHash) &&
			c.expectedReturn.FirstName != user.FirstName && c.expectedReturn.LastName != user.LastName {
			t.Errorf("case %s: incorrect return: expected %v but revieved %v",
				c.name, c.expectedReturn, user)
		}

	}

}

func TestSpecificSessionHandler(t *testing.T) {

	fakeConn := users.NewFakeConnection()
	fakeSession := sessions.NewMemStore(defaultSessionDuration, defaultSessionDuration)

	var handleTest = HandlerContext{
		Key:     "fakesigningkey",
		User:    fakeConn,
		Session: fakeSession,
	}

	validNewUser := users.NewUser{
		Email:        "coolguy@gmail.com",
		Password:     "password",
		PasswordConf: "password",
		UserName:     "coolguy",
		FirstName:    "Cool",
		LastName:     "Guy",
	}

	validUser, _ := validNewUser.ToUser()

	cases := []struct {
		name               string
		method             string
		requestContentType string
		validUser          *users.User
		expectedStatusCode int
		expectedError      bool
		url                string
	}{

		{
			"Valid Delete Request",
			http.MethodDelete,
			HeaderJSON,
			validUser,
			http.StatusOK,
			false,
			"/v1/sessions/mine",
		},

		{
			"Non Delete",
			http.MethodGet,
			HeaderJSON,
			validUser,
			http.StatusOK,
			true,
			"/v1/sessions/mine",
		},

		{
			"Invalid",
			http.MethodGet,
			HeaderJSON,
			validUser,
			http.StatusForbidden,
			true,
			"/v1/sessions/notmine",
		},
	}

	for _, c := range cases {

		request := httptest.NewRequest(c.method, c.url, nil)
		recorder := httptest.NewRecorder()

		var sessionState UserSession
		sessionState.User = *validUser
		sessionState.Usertime = time.Now()

		sid, _ := sessions.BeginSession(handleTest.Key, handleTest.Session, &sessionState, recorder)

		var testState UserSession
		err := handleTest.Session.Get(sid, &testState)

		if c.validUser.ID != testState.User.ID || err != nil {
			t.Errorf("The sessionstate was not initalized properly")
		}

		request.Header.Add(HeaderAuth, recorder.Header().Get(HeaderAuth))
		handleTest.SpecificSessionHandler(recorder, request)

		response := recorder.Result()

		resStatusCode := response.StatusCode
		if c.expectedStatusCode != resStatusCode {
			t.Errorf("case %s: incorrect status code: expected: %d recieved: %d",
				c.name, c.expectedStatusCode, resStatusCode)
		}

		//Check that it is deleted
		testState = UserSession{}
		err = handleTest.Session.Get(sid, &testState)

		if err == nil {
			t.Errorf("Session was not removed properly")
		}

	}

}
