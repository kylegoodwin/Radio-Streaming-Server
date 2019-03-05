package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/assignments-zanewebbUW/servers/gateway/models/users"
	"github.com/assignments-zanewebbUW/servers/gateway/sessions"
	"github.com/go-redis/redis"
)

func TestContext(t *testing.T) {
	//Fail Case
	var a *redis.Client
	b := sessions.NewRedisStore(a, time.Second) // sessions.NewMemStore(time.Hour, time.Hour)
	c := &users.MySQLStore{}                    //&users.MyMockStore{}
	context := NewContext("", b, c)
	if context != nil {
		t.Error("Expected Context constructor to fail but it did not return nil")
	}

	//Success Case
	a = redis.NewClient(&redis.Options{
		Addr: "172.17.0.2:6379",
	})
	b = sessions.NewRedisStore(a, time.Second)
	//b = sessions.NewMemStore(time.Hour, time.Hour)
	context = NewContext("test", b, c)
	if context == nil {
		t.Error("Expected Context constructor to work but it did not ")
	}
}

func TestContext_POSTUserHandler(t *testing.T) {
	//make sure to run this badboy:
	//sudo docker run -d --name redisServer redis
	//and set your env variable
	redisaddr := os.Getenv("REDISADDR")
	if len(redisaddr) == 0 {
		redisaddr = "172.17.0.2:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisaddr,
	})

	context := &Context{
		Key:           "testkey",
		SessionsStore: sessions.NewRedisStore(client, time.Hour), //sessions.NewMemStore(time.Hour, time.Hour),
		UsersStore:    &users.MySQLStore{},                       //&users.MyMockStore{},
	}

	//CREATE NEW USERS FOR TEST CASES
	/*
		Email        string `json:"email"`
		Password     string `json:"password"`
		PasswordConf string `json:"passwordConf"`
		UserName     string `json:"userName"`
		FirstName    string `json:"firstName"`
		LastName     string `json:"lastName"`
	*/
	validNewUser := users.NewUser{
		Email:        "test@test.com",
		Password:     "hunter2",
		PasswordConf: "hunter2",
		UserName:     "AzureDiamond",
		FirstName:    "John",
		LastName:     "Johnson",
	}

	invalidNewUser := users.NewUser{
		Email:        "test@test.com",
		Password:     "hunter2",
		PasswordConf: "Wrong",
		UserName:     "AzureDiamond",
		FirstName:    "Error",
		LastName:     "Johnson",
	}
	FailInsertNewUser1 := users.NewUser{
		Email:        "test@test.com",
		Password:     "hunter2",
		PasswordConf: "hunter2",
		UserName:     "AzureDiamond",
		FirstName:    "Error",
		LastName:     "Johnson",
	}
	FailInsertNewUser2 := users.NewUser{
		Email:        "test@test.com",
		Password:     "hunter2",
		PasswordConf: "hunter2",
		UserName:     "AzureDiamond",
		FirstName:    "Error",
		LastName:     "Error",
	}

	//CREATE USERS FOR TEST CASE RESULTS
	validUser, err := validNewUser.ToUser()
	if err != nil {
		log.Printf("error initializing test users")
	}
	validUser.ID = int64(1)

	/*invalidUser, err := invalidNewUser.ToUser()
	if err != nil {
		log.Printf("error initializing test users")
	}*/

	cases := []struct {
		name                string
		method              string
		idPath              string
		newUser             *users.NewUser
		expectedStatusCode  int
		expectedError       bool
		expectedContentType string
		expectedReturn      *users.User
	}{
		{
			"Valid Post Request",
			http.MethodPost,
			"1",
			&validNewUser,
			http.StatusCreated,
			false,
			"application/json",
			validUser,
		},
		{
			"Invalid Post Request - Content Type",
			http.MethodPost,
			"2",
			&validNewUser,
			http.StatusUnsupportedMediaType,
			true,
			"text/plain; charset=utf-8",
			validUser,
		},
		{
			"Invalid Post Request - Method",
			http.MethodGet,
			"2",
			&validNewUser,
			http.StatusMethodNotAllowed,
			true,
			"text/plain; charset=utf-8",
			validUser,
		},
		{
			"Invalid Post Request - Invalid NewUser",
			http.MethodPost,
			"2",
			&invalidNewUser,
			http.StatusInternalServerError,
			true,
			"text/plain; charset=utf-8",
			validUser,
		},
		{
			"Invalid Given User - Insert Fail",
			http.MethodPost,
			"2",
			&FailInsertNewUser1,
			http.StatusInternalServerError,
			true,
			"text/plain; charset=utf-8",
			validUser,
		},
		{
			"Invalid Given User - Insert Fail (ID)",
			http.MethodPost,
			"2",
			&FailInsertNewUser2,
			http.StatusInternalServerError,
			true,
			"text/plain; charset=utf-8",
			validUser,
		},
	}

	for _, c := range cases {
		log.Printf("running case name: %s", c.name)
		buffer := new(bytes.Buffer)
		encoder := json.NewEncoder(buffer)
		encoder.Encode(c.newUser)
		request := httptest.NewRequest(c.method, "/v1/users/"+c.idPath, buffer)

		if c.expectedStatusCode != http.StatusUnsupportedMediaType {
			request.Header.Add("Content-Type", "application/json")
		}

		recorder := httptest.NewRecorder()
		context.UsersHandler(recorder, request)
		response := recorder.Result()

		resContentType := response.Header.Get("Content-Type")
		if !c.expectedError && c.expectedContentType != resContentType {
			t.Errorf("case %s: incorrect return type: expected: %s received: %s",
				c.name, c.expectedContentType, resContentType)
		}

		resStatusCode := response.StatusCode
		if c.expectedStatusCode != resStatusCode {
			t.Errorf("case %s: incorrect status code: expected: %d received: %d",
				c.name, c.expectedStatusCode, resStatusCode)
		}

		user := &users.User{}
		err := json.NewDecoder(response.Body).Decode(user)
		if c.expectedError && err == nil {
			t.Errorf("case %s: expected error but received none", c.name)
		}

		if !c.expectedError && c.expectedReturn.Email != user.Email && string(c.expectedReturn.PassHash) != string(user.PassHash) &&
			c.expectedReturn.FirstName != user.FirstName && c.expectedReturn.LastName != user.LastName {
			t.Errorf("case %s: incorrect return: expected %v but received %v",
				c.name, c.expectedReturn, user)
		}
	}

}

func TestContext_GETSpecificUserHandler(t *testing.T) {
	redisaddr := os.Getenv("REDISADDR")
	if len(redisaddr) == 0 {
		redisaddr = "172.17.0.2:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisaddr,
	})

	context := &Context{
		Key:           "testkey",
		SessionsStore: sessions.NewRedisStore(client, time.Hour), //sessions.NewMemStore(time.Hour, time.Hour),
		UsersStore:    &users.MySQLStore{},                       //&users.MyMockStore{},
	}

	validNewUser := users.NewUser{
		Email:        "test@test.com",
		Password:     "hunter2",
		PasswordConf: "hunter2",
		UserName:     "AzureDiamond",
		FirstName:    "John",
		LastName:     "Johnson",
	}
	/*invalidNewUser := users.NewUser{
		Email:        "wrong@test.com",
		Password:     "hunter2",
		PasswordConf: "hunter2",
		UserName:     "AzureDiamond",
		FirstName:    "John",
		LastName:     "Johnson",
	}*/

	//CREATE USERS FOR TEST CASE RESULTS
	validUser, err := validNewUser.ToUser()
	if err != nil {
		log.Printf("error initializing test users")
	}
	validUser.ID = int64(1)

	cases := []struct {
		name                string
		method              string
		idPath              string
		newUser             *users.NewUser
		useValidCredentials bool
		expectedStatusCode  int
		expectedError       bool
		expectedContentType string
		expectedReturn      *users.User
	}{
		{
			"Valid Get Request",
			http.MethodGet,
			"1",
			&validNewUser,
			true,
			http.StatusOK,
			false,
			"application/json",
			validUser,
		},
		{
			"Valid Get Request",
			http.MethodGet,
			"me",
			&validNewUser,
			true,
			http.StatusOK,
			false,
			"application/json",
			validUser,
		},
		{
			"Invalid Get Request - No Credentials (no session)",
			http.MethodGet,
			"1",
			&validNewUser,
			false,
			http.StatusUnauthorized,
			true,
			"text/plain; charset=utf-8",
			validUser,
		},
		{
			"Invalid Get Request - Wrong Method",
			http.MethodPost,
			"1",
			&validNewUser,
			true,
			http.StatusMethodNotAllowed,
			true,
			"text/plain; charset=utf-8",
			validUser,
		},
		{
			"Invalid Get Request - Requested User Not Found",
			http.MethodGet,
			"2",
			&validNewUser,
			true,
			http.StatusNotFound,
			true,
			"text/plain; charset=utf-8",
			validUser,
		},
		{
			"Invalid Get Request - No Provided ID",
			http.MethodGet,
			"",
			&validNewUser,
			true,
			http.StatusNotFound,
			true,
			"text/plain; charset=utf-8",
			validUser,
		},
	}

	for _, c := range cases {

		log.Printf("\n\nrunning case name: %s", c.name)
		buffer := new(bytes.Buffer)
		encoder := json.NewEncoder(buffer)
		encoder.Encode(c.newUser)
		request := httptest.NewRequest(c.method, "/v1/users/"+c.idPath, buffer)

		if c.expectedStatusCode != http.StatusUnsupportedMediaType {
			request.Header.Add("Content-Type", "application/json")
		}

		recorder := httptest.NewRecorder()
		//Trying to create a session for the valid user
		if c.name != "Invalid Get Request - No Credentials (no session)" {
			newSessState := SessionState{time.Now(), *c.expectedReturn}
			sid, err := sessions.BeginSession(context.Key, context.SessionsStore, &newSessState, recorder)
			if err != nil {
				log.Fatal(err.Error())
			}
			request.Header.Add("Authorization", "Bearer "+sid.String())

		}

		context.SpecificUserHandler(recorder, request)
		response := recorder.Result()

		resContentType := response.Header.Get("Content-Type")
		if !c.expectedError && c.expectedContentType != resContentType {
			t.Errorf("case %s: incorrect return type: expected: %s received: %s",
				c.name, c.expectedContentType, resContentType)
		}

		resStatusCode := response.StatusCode
		if c.expectedStatusCode != resStatusCode {
			t.Errorf("case %s: incorrect status code: expected: %d received: %d",
				c.name, c.expectedStatusCode, resStatusCode)
		}

		user := &users.User{}
		err := json.NewDecoder(response.Body).Decode(user)
		if c.expectedError && err == nil {
			t.Errorf("case %s: expected error but received none", c.name)
		}

		if !c.expectedError && c.expectedReturn.Email != user.Email && string(c.expectedReturn.PassHash) != string(user.PassHash) &&
			c.expectedReturn.FirstName != user.FirstName && c.expectedReturn.LastName != user.LastName {
			t.Errorf("case %s: incorrect return: expected %v but received %v",
				c.name, c.expectedReturn, user)
		}
	}
}
func TestContext_PATCHSpecificUserHandler(t *testing.T) {
	redisaddr := os.Getenv("REDISADDR")
	if len(redisaddr) == 0 {
		redisaddr = "172.17.0.2:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisaddr,
	})

	context := &Context{
		Key:           "testkey",
		SessionsStore: sessions.NewRedisStore(client, time.Hour), //sessions.NewMemStore(time.Hour, time.Hour),
		UsersStore:    &users.MySQLStore{},                       //&users.MyMockStore{},
	}

	validNewUser := users.NewUser{
		Email:        "test@test.com",
		Password:     "hunter2",
		PasswordConf: "hunter2",
		UserName:     "AzureDiamond",
		FirstName:    "John",
		LastName:     "Johnson",
	}

	//CREATE USERS FOR TEST CASE RESULTS
	validUser, err := validNewUser.ToUser()
	if err != nil {
		log.Printf("error initializing test users")
	}
	validUser.ID = int64(1)

	//CREATE UPDATES FOR TEST CASES
	validUpdates := users.Updates{
		FirstName: "Test",
		LastName:  "Johnson",
	}
	invalidUpdates := users.Updates{
		FirstName: "Error",
		LastName:  "Johnson",
	}

	cases := []struct {
		name                string
		method              string
		idPath              string
		updates             *users.Updates
		useValidCredentials bool
		expectedStatusCode  int
		expectedError       bool
		expectedContentType string
		expectedReturn      *users.User
	}{
		{
			"Valid Patch Request",
			http.MethodPatch,
			"1",
			&validUpdates,
			true,
			http.StatusOK,
			false,
			"application/json",
			validUser,
		},
		{
			"Invalid Patch Request - Invalid upadates",
			http.MethodPatch,
			"1",
			&invalidUpdates,
			true,
			http.StatusInternalServerError,
			true,
			"text/plain; charset=utf-8",
			validUser,
		},
		{
			"Invalid Patch Request - Trying to update another user",
			http.MethodPatch,
			"2",
			&invalidUpdates,
			true,
			http.StatusForbidden,
			true,
			"text/plain; charset=utf-8",
			validUser,
		},
		{
			"Invalid Patch Request - No ID given",
			http.MethodPatch,
			"",
			&invalidUpdates,
			true,
			http.StatusNotFound,
			true,
			"text/plain; charset=utf-8",
			validUser,
		},
		{
			"Invalid Patch Request - Invalid header",
			http.MethodPatch,
			"1",
			&invalidUpdates,
			true,
			http.StatusUnsupportedMediaType,
			true,
			"text/plain; charset=utf-8",
			validUser,
		},
	}

	for _, c := range cases {

		log.Printf("\n\nrunning case name: %s", c.name)
		buffer := new(bytes.Buffer)
		encoder := json.NewEncoder(buffer)
		encoder.Encode(c.updates)
		request := httptest.NewRequest(c.method, "/v1/users/"+c.idPath, buffer)

		if c.expectedStatusCode != http.StatusUnsupportedMediaType {
			request.Header.Add("Content-Type", "application/json")
		}

		recorder := httptest.NewRecorder()
		//Trying to create a session for the valid user
		if c.name != "Invalid Get Request - No Credentials (no session)" {
			newSessState := SessionState{time.Now(), *c.expectedReturn}
			sid, err := sessions.BeginSession(context.Key, context.SessionsStore, &newSessState, recorder)
			if err != nil {
				log.Fatal(err.Error())
			}
			request.Header.Add("Authorization", "Bearer "+sid.String())

		}

		context.SpecificUserHandler(recorder, request)
		response := recorder.Result()

		resContentType := response.Header.Get("Content-Type")
		if !c.expectedError && c.expectedContentType != resContentType {
			t.Errorf("case %s: incorrect return type: expected: %s received: %s",
				c.name, c.expectedContentType, resContentType)
		}

		resStatusCode := response.StatusCode
		if c.expectedStatusCode != resStatusCode {
			t.Errorf("case %s: incorrect status code: expected: %d received: %d",
				c.name, c.expectedStatusCode, resStatusCode)
		}

		user := &users.User{}
		err := json.NewDecoder(response.Body).Decode(user)
		if c.expectedError && err == nil {
			t.Errorf("case %s: expected error but received none", c.name)
		}

		if !c.expectedError && c.expectedReturn.Email != user.Email && string(c.expectedReturn.PassHash) != string(user.PassHash) &&
			c.expectedReturn.FirstName != user.FirstName && c.expectedReturn.LastName != user.LastName {
			t.Errorf("case %s: incorrect return: expected %v but received %v",
				c.name, c.expectedReturn, user)
		}
	}
}

func TestContext_SessionsHandler(t *testing.T) {
	redisaddr := os.Getenv("REDISADDR")
	if len(redisaddr) == 0 {
		redisaddr = "172.17.0.2:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisaddr,
	})

	context := &Context{
		Key:           "testkey",
		SessionsStore: sessions.NewRedisStore(client, time.Hour), // sessions.NewMemStore(time.Hour, time.Hour),
		UsersStore:    &users.MySQLStore{},                       //&users.MyMockStore{},
	}

	//CREATE CREDENTIALS FOR TEST CASES
	validCredentials := users.Credentials{
		Email:    "test@test.com",
		Password: "hunter2",
	}

	invalidCredentialsPass := users.Credentials{
		Email:    "test@test.com",
		Password: "wrong",
	}
	invalidCredentialsEmail := users.Credentials{
		Email:    "wrong@test.com",
		Password: "hunter2",
	}
	var malformedCredentials *users.Credentials

	//CREATE NEW USERS FOR TEST CASES
	/*
		Email        string `json:"email"`
		Password     string `json:"password"`
		PasswordConf string `json:"passwordConf"`
		UserName     string `json:"userName"`
		FirstName    string `json:"firstName"`
		LastName     string `json:"lastName"`
	*/
	validNewUser := users.NewUser{
		Email:        "test@test.com",
		Password:     "hunter2",
		PasswordConf: "hunter2",
		UserName:     "AzureDiamond",
		FirstName:    "John",
		LastName:     "Johnson",
	}

	//CREATE USERS FOR TEST CASE RESULTS
	validUser, err := validNewUser.ToUser()
	if err != nil {
		log.Printf("error initializing test users")
	}
	validUser.ID = int64(1)

	//CREATE CASES
	cases := []struct {
		name                string
		method              string
		idPath              string
		credentials         *users.Credentials
		expectedStatusCode  int
		expectedError       bool
		expectedContentType string
		expectedReturn      *users.User
	}{
		{
			"Valid Post Request",
			http.MethodPost,
			"1",
			&validCredentials,
			http.StatusOK,
			false,
			"application/json",
			validUser,
		},
		{
			"Valid Post Request - Wrong Credentials Email",
			http.MethodPost,
			"1",
			&invalidCredentialsEmail,
			http.StatusUnauthorized,
			true,
			"text/plain; charset=utf-8",
			validUser,
		},
		{
			"Valid Post Request - Wrong Credentials Pass",
			http.MethodPost,
			"1",
			&invalidCredentialsPass,
			http.StatusUnauthorized,
			true,
			"text/plain; charset=utf-8",
			validUser,
		},
		{
			"Invalid Post Request - Wrong Method",
			http.MethodGet,
			"1",
			&validCredentials,
			http.StatusMethodNotAllowed,
			true,
			"text/plain; charset=utf-8",
			validUser,
		},
		{
			"Invalid Post Request - Wrong Header",
			http.MethodPost,
			"1",
			&validCredentials,
			http.StatusUnsupportedMediaType,
			true,
			"text/plain; charset=utf-8",
			validUser,
		},
		{
			"Invalid Post Request - Malformed Credentials",
			http.MethodPost,
			"1",
			malformedCredentials,
			http.StatusInternalServerError,
			true,
			"text/plain; charset=utf-8",
			validUser,
		},
	}

	//Generate and store new session for valid test
	//newSessState := SessionState{time.Now(), *validUser}
	//sessions.BeginSession()

	for _, c := range cases {
		log.Printf("running case name: %s", c.name)
		buffer := new(bytes.Buffer)
		encoder := json.NewEncoder(buffer)
		if c.credentials != malformedCredentials {
			encoder.Encode(c.credentials)
		}
		request := httptest.NewRequest(c.method, "/v1/users/"+c.idPath, buffer)

		if c.expectedStatusCode != http.StatusUnsupportedMediaType {
			request.Header.Add("Content-Type", "application/json")
		}
		// if c.expectedStatusCode != http.StatusUnauthorized {
		// 	request.Header.Add("Authorization", "Bearer "+sid.String())
		// }

		recorder := httptest.NewRecorder()
		context.SessionsHandler(recorder, request)
		response := recorder.Result()

		resContentType := response.Header.Get("Content-Type")
		if !c.expectedError && c.expectedContentType != resContentType {
			t.Errorf("case %s: incorrect return type: expected: %s received: %s",
				c.name, c.expectedContentType, resContentType)
		}

		resStatusCode := response.StatusCode
		if c.expectedStatusCode != resStatusCode {
			t.Errorf("case %s: incorrect status code: expected: %d received: %d",
				c.name, c.expectedStatusCode, resStatusCode)
		}

		user := &users.User{}
		err := json.NewDecoder(response.Body).Decode(user)
		if c.expectedError && err == nil {
			t.Errorf("case %s: expected error but received none", c.name)
		}

		if !c.expectedError && c.expectedReturn.Email != user.Email && string(c.expectedReturn.PassHash) != string(user.PassHash) &&
			c.expectedReturn.FirstName != user.FirstName && c.expectedReturn.LastName != user.LastName {
			t.Errorf("case %s: incorrect return: expected %v but received %v",
				c.name, c.expectedReturn, user)
		}
	}
}

func TestContext_DELETESpecificSessionHandler(t *testing.T) {
	redisaddr := os.Getenv("REDISADDR")
	if len(redisaddr) == 0 {
		redisaddr = "172.17.0.2:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisaddr,
	})

	context := &Context{
		Key:           "testkey",
		SessionsStore: sessions.NewRedisStore(client, time.Hour), //sessions.NewMemStore(time.Hour, time.Hour),
		UsersStore:    &users.MySQLStore{},                       //&users.MyMockStore{},
	}

	validNewUser := users.NewUser{
		Email:        "test@test.com",
		Password:     "hunter2",
		PasswordConf: "hunter2",
		UserName:     "AzureDiamond",
		FirstName:    "John",
		LastName:     "Johnson",
	}

	//CREATE USERS FOR TEST CASE RESULTS
	validUser, err := validNewUser.ToUser()
	if err != nil {
		log.Printf("error initializing test users")
	}
	validUser.ID = int64(1)

	cases := []struct {
		name                string
		method              string
		idPath              string
		useValidCredentials bool
		expectedStatusCode  int
		expectedError       bool
		expectedContentType string
		expectedReturn      string
	}{
		{
			"Valid Delete Request",
			http.MethodDelete,
			"mine",
			true,
			http.StatusOK,
			false,
			"text/plain; charset=utf-8",
			"signed out",
		},
		{
			"Invalid Delete Request - Invalid Method",
			http.MethodPatch,
			"1",
			true,
			http.StatusMethodNotAllowed,
			true,
			"text/plain; charset=utf-8",
			"Expected Error",
		},
		{
			"Invalid Delete Request - Trying to log out another user",
			http.MethodDelete,
			"2",
			true,
			http.StatusForbidden,
			true,
			"text/plain; charset=utf-8",
			"Expected Error",
		},
		{
			"Invalid Delete Request - No existing session",
			http.MethodDelete,
			"2",
			false,
			http.StatusUnauthorized,
			true,
			"text/plain; charset=utf-8",
			"Expected Error",
		},
	}

	for _, c := range cases {

		log.Printf("\n\nrunning case name: %s", c.name)
		request := httptest.NewRequest(c.method, "/v1/sessions/"+c.idPath, nil)

		if c.expectedStatusCode != http.StatusUnsupportedMediaType {
			request.Header.Add("Content-Type", "application/json")
		}

		recorder := httptest.NewRecorder()
		//Trying to create a session for the valid user
		if c.useValidCredentials {
			newSessState := SessionState{time.Now(), *validUser}
			sid, err := sessions.BeginSession(context.Key, context.SessionsStore, &newSessState, recorder)
			if err != nil {
				log.Fatal(err.Error())
			}
			request.Header.Add("Authorization", "Bearer "+sid.String())

		}

		context.SpecificSessionHandler(recorder, request)
		response := recorder.Result()

		resContentType := response.Header.Get("Content-Type")
		if !c.expectedError && c.expectedContentType != resContentType {
			t.Errorf("case %s: incorrect return type: expected: %s received: %s",
				c.name, c.expectedContentType, resContentType)
		}

		resStatusCode := response.StatusCode
		if c.expectedStatusCode != resStatusCode {
			t.Errorf("case %s: incorrect status code: expected: %d received: %d",
				c.name, c.expectedStatusCode, resStatusCode)
		}

		// user := &users.User{}
		// err := json.NewDecoder(response.Body).Decode(user)
		// if c.expectedError && err == nil {
		// 	t.Errorf("case %s: expected error but received none", c.name)
		// }
		responseData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			t.Errorf("Crashed trying to read response data")
		}
		responseString := string(responseData)
		log.Printf("Returned string was: %s", responseString)

		if !c.expectedError && c.expectedReturn != "signed out" {
			t.Errorf("case %s: incorrect return: expected %v but received %v",
				c.name, c.expectedReturn, responseString)
		}
	}
}
