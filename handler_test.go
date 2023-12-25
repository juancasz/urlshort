package urlshort

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

	t.Run("method not GET not valid", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello, world!")
		})
		handler := MapHandler(map[string]string{
			"/urlshort":       "https://github.com/gophercises/urlshort",
			"/urlshort-final": "https://github.com/gophercises/urlshort/tree/solution",
		}, mux)
		req, err := http.NewRequest("POST", "/urlshort", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		handler(rr, req)
		if status := rr.Code; status != http.StatusMethodNotAllowed {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
		}
	})
}

func TestYAMLHandler(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world!")
	})

	tests := map[string]struct {
		yml                []byte
		fallback           http.Handler
		requestPath        string
		expectedStatusCode int
	}{
		"correct handler": {
			yml: []byte(`
- path: /urlshort
  url: https://github.com/gophercises/urlshort
- path: /urlshort-final
  url: https://github.com/gophercises/urlshort/tree/solution
`),
			fallback:           mux,
			requestPath:        "/urlshort-final",
			expectedStatusCode: 301,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			handler, err := YAMLHandler(tc.yml, tc.fallback)
			if err != nil {
				t.Fatal(err)
			}
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

func TestJSONHandler(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world!")
	})

	tests := map[string]struct {
		json               []byte
		fallback           http.Handler
		requestPath        string
		expectedStatusCode int
	}{
		"correct handler": {
			json: []byte(`
[
	{
		"path": "/urlshort",
		"url": "https://github.com/gophercises/urlshort"
	},
	{
		"path": "/urlshort-final",
		"url": "https://github.com/gophercises/urlshort/tree/solution"
	}
]
`),
			fallback:           mux,
			requestPath:        "/urlshort-final",
			expectedStatusCode: 301,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			handler, err := JSONHandler(tc.json, tc.fallback)
			if err != nil {
				t.Fatal(err)
			}
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

type mockSaver struct{}

func (m *mockSaver) Save(ctx context.Context, key string, url string) error {
	return nil
}

type mockSaverError struct{}

func (m *mockSaverError) Save(ctx context.Context, key string, url string) error {
	return fmt.Errorf("something happened")
}

func TestShortener(t *testing.T) {
	tests := map[string]struct {
		URL        string
		Method     string
		response   string
		statusCode int
	}{
		"valid URL with scheme": {
			URL:        "http://google.com",
			Method:     "POST",
			statusCode: http.StatusOK,
		},
		"valid URL full": {
			URL:        "http://www.google.com",
			Method:     "POST",
			statusCode: http.StatusOK,
		},
		"invalid URL": {
			URL:        "google",
			Method:     "POST",
			response:   "invalid URL",
			statusCode: http.StatusBadRequest,
		},
		"invalid URL relative path": {
			URL:        "google.com",
			Method:     "POST",
			response:   "invalid URL",
			statusCode: http.StatusBadRequest,
		},
		"empty URL": {
			URL:        "",
			Method:     "POST",
			response:   "URL parameter is missing",
			statusCode: http.StatusBadRequest,
		},
		"invalid method": {
			URL:        "http://www.google.com",
			Method:     "GET",
			response:   "Invalid request method",
			statusCode: http.StatusMethodNotAllowed,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			handler := Shortener(&mockSaver{})
			req, err := http.NewRequest(tc.Method, fmt.Sprintf("/shorten?url=%s", tc.URL), nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler(rr, req)
			if status := rr.Code; status != tc.statusCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.statusCode)
			}
			if rr.Code != http.StatusOK {
				if strings.TrimSpace(rr.Body.String()) != tc.response {
					t.Errorf("handler returned unexpected body: got %v want %v",
						rr.Body.String(), tc.response)
				}
			}
		})
	}

	t.Run("error saving url shortened", func(t *testing.T) {
		handler := Shortener(&mockSaverError{})
		req, err := http.NewRequest("POST", fmt.Sprintf("/shorten?url=%s", "http://www.google.com"), nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		handler(rr, req)
		if status := rr.Code; status != http.StatusInternalServerError {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
		}
	})
}
