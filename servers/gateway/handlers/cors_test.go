package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/assignments-zanewebbUW/servers/gateway/handlers"
)

func TestCors(t *testing.T) {
	var nilHandler http.Handler

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/v1/users/", nil)

	var testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("IWDLAMWLDN"))
	})

	cors := handlers.NewCors(nilHandler)
	if cors != nil {
		t.Errorf("Expected constructor to return nil but it did not")
	}

	cors = handlers.NewCors(testHandler)
	if cors == nil {
		t.Errorf("Expected constructor to not return nil but it did")
	}

	cors.ServeHTTP(recorder, request)

	response := recorder.Result()

	if response.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Header 1 did not get added")
	}
	if response.Header.Get("Access-Control-Allow-Methods") != "GET, PUT, POST, PATCH, DELETE" {
		t.Errorf("Header 2 did not get added")
	}
	if response.Header.Get("Access-Control-Allow-Headers") != "Content-Type, Authorization" {
		t.Errorf("Header 3 did not get added")
	}
	if response.Header.Get("Access-Control-Expose-Headers") != "Authorization" {
		t.Errorf("Header 4 did not get added")
	}
	if response.Header.Get("Access-Control-Max-Age") != "600" {
		t.Errorf("Header 5 did not get added")
	}

}
