package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCors(t *testing.T) {

	request := httptest.NewRequest("GET", "/v1/users", nil)
	recorder := httptest.NewRecorder()

	var testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("bar"))
	})

	cors := Cors{}
	cors.Handler = testHandler

	cors.ServeHTTP(recorder, request)

	response := recorder.Result()

	fmt.Println(response)

	if response.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin to equal * but got %s", response.Header.Get("Access-Control-Allow-Origin"))
	}

	if response.Header.Get("Access-Control-Allow-Methods") != "GET, PUT, POST, PATCH, DELETE" {
		t.Errorf("Expected Access-Control-Allow-Methods to equal GET, PUT, POST, PATCH, DELETE but got %s", response.Header.Get("Access-Control-Allow-Methods"))
	}

	if response.Header.Get("Access-Control-Allow-Headers") != "Content-Type, Authorization" {
		t.Errorf("Expected Access-Control-Allow-Headers to equal Content-Type, Authorization but got %s", response.Header.Get("Access-Control-Allow-Headers"))
	}

	if response.Header.Get("Access-Control-Expose-Headers") != "Authorization" {
		t.Errorf("Expected Access-Control-Expose-Headers to equal Authorization but got %s", response.Header.Get("Access-Control-Expose-Headers"))
	}

	if response.Header.Get("Access-Control-Max-Age") != "600" {
		t.Errorf("Expected Access-Control-Max-Age to equal 600 but got %s", response.Header.Get("Access-Control-Max-Age"))
	}

}
