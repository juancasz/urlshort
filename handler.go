// Package urlshort provides functions to implement a small server with urlshortener functionalities
package urlshort

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
		if redirectUrl, ok := pathsToUrls[strings.TrimRight(r.URL.Path, "/ ")]; ok {
			http.Redirect(w, r, redirectUrl, http.StatusMovedPermanently)
			return
		}
		fallback.ServeHTTP(w, r)
	}
}

type uRLMapper struct {
	Path string `yaml:"path" json:"path"`
	URL  string `yaml:"url" json:"url"`
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//   - path: /some-path
//     url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func YAMLHandler(yml []byte, fallback http.Handler) (http.HandlerFunc, error) {
	var mappers []uRLMapper
	err := yaml.Unmarshal(yml, &mappers)
	if err != nil {
		return nil, err
	}
	pathMap, err := buildMap(mappers)
	if err != nil {
		return nil, err
	}
	return MapHandler(pathMap, fallback), nil
}

// JSONHandler will parse the provided JSON and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// JSON is expected to be in the format:
//
// [
//
//	{
//	    "path": "/urlshort",
//	    "url": "https://github.com/gophercises/urlshort"
//	},
//	{
//	    "path": "/urlshort-final",
//	    "url": "https://github.com/gophercises/urlshort/tree/solution"
//	}
//
// ]
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func JSONHandler(data []byte, fallback http.Handler) (http.HandlerFunc, error) {
	var mappers []uRLMapper
	err := json.Unmarshal(data, &mappers)
	if err != nil {
		return nil, err
	}
	pathMap, err := buildMap(mappers)
	if err != nil {
		return nil, err
	}
	return MapHandler(pathMap, fallback), nil
}

func buildMap(mappers []uRLMapper) (map[string]string, error) {
	mapOutput := make(map[string]string)
	var ok bool
	for _, mapper := range mappers {
		if _, ok = mapOutput[mapper.Path]; ok {
			return nil, fmt.Errorf("repeated path")
		}
		mapOutput[mapper.Path] = mapper.URL
	}
	return mapOutput, nil
}

// UrlShortSaver defines a contract for types that know how to save a shortened URL key.
// Types implementing this interface must provide a Save method that takes a string representing the key
// and returns an error if the operation fails.
type UrlShortSaver interface {
	// Save is a method that takes a string key representing the URL to be saved.
	// It attempts to save the key and returns an error if the operation fails.
	Save(ctx context.Context, key string, url string) error
}

// Shortener generates an HTTP handler that accepts POST requests containing a URL.
// It then generates a shortened key for the provided URL and saves it using the provided saver.
// The generated shortened URL is displayed in the HTML response along with the original URL.
func Shortener(saver UrlShortSaver, host string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		originalURL := r.FormValue("url")
		if originalURL == "" {
			http.Error(w, "URL parameter is missing", http.StatusBadRequest)
			return
		}

		_, err := url.ParseRequestURI(originalURL)
		if err != nil {
			fallback.ServeHTTP(w, r)
			return
		}

		shortKey := generateShortKey()
		if err = saver.Save(r.Context(), shortKey, originalURL); err != nil {
			http.Error(w, fmt.Sprintf("error saving short url"), http.StatusInternalServerError)
			return
		}

		shortenedURL := fmt.Sprintf("%s/short/%s", host, shortKey)

		tmpl, err := template.ParseFiles("html/shorten.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err = tmpl.Execute(w, struct {
			OriginalUrl string
			ShortUrl    string
		}{
			OriginalUrl: originalURL,
			ShortUrl:    shortenedURL,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func generateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 6

	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)
	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rng.Intn(len(charset))]
	}
	return string(shortKey)
}

// UrlShortGetter defines a contract for types that know how to redirect a shortened URL key.
// Types implementing this interface must provide a Get method that takes a string representing the key
// and returns the url to which some request must be redirected.
type UrlShortGetter interface {
	// Get is a method that takes a string key representing the shortened URL and returns the original url
	// to which some request must be redirected
	Get(ctx context.Context, key string) (string, error)
}

var ErrMissingKey = errors.New("key not found")

// RetrieveHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to redirect any
// paths (keys) to their corresponding URL (values
// that UrlShortGetter retrieves, in string format).
// If the key is not found by getter, then the fallback
// http.Handler will be called instead.
// Handler must be attached to route /anypath/{key} or it won't work properly
func RetrieveHandler(getter UrlShortGetter, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
		paths := strings.SplitN(strings.Trim(r.URL.Path, "/ "), "/", 2)
		if len(paths) != 2 {
			http.NotFound(w, r)
			return
		}
		redirectUrl, err := getter.Get(r.Context(), paths[1])
		if errors.Is(err, ErrMissingKey) {
			fallback.ServeHTTP(w, r)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, redirectUrl, http.StatusMovedPermanently)
	}
}

// ShortenerHome returns home page for shortener website
func ShortenerHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := template.ParseFiles("html/home.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// MissingUrlHandler returns page when key not found
func MissingUrlHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/fallback.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// InvalidUrlHandler returns page when url not valid
func InvalidUrlHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/error.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
