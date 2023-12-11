package urlshort

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMapHandler(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world!")
	})

	tests := map[string]struct {
		pathsToUrls        map[string]string
		fallback           http.Handler
		requestPath        string
		expectedStatusCode int
	}{
		"correct handler": {
			pathsToUrls: map[string]string{
				"/urlshort":       "https://github.com/gophercises/urlshort",
				"/urlshort-final": "https://github.com/gophercises/urlshort/tree/solution",
			},
			fallback:           mux,
			requestPath:        "/urlshort",
			expectedStatusCode: 301,
		},
		"correct handler trailing space and slash": {
			pathsToUrls: map[string]string{
				"/urlshort":       "https://github.com/gophercises/urlshort",
				"/urlshort-final": "https://github.com/gophercises/urlshort/tree/solution",
			},
			fallback:           mux,
			requestPath:        "/urlshort-final/  ",
			expectedStatusCode: 301,
		},
		"not mapped path": {
			pathsToUrls: map[string]string{
				"/urlshort":       "https://github.com/gophercises/urlshort",
				"/urlshort-final": "https://github.com/gophercises/urlshort/tree/solution",
			},
			fallback:           mux,
			requestPath:        "/random-path",
			expectedStatusCode: 200,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			handler := MapHandler(tc.pathsToUrls, tc.fallback)
			req, err := http.NewRequest("GET", tc.requestPath, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler(rr, req)
			if status := rr.Code; status != tc.expectedStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.expectedStatusCode)
			}
		})
	}
}
