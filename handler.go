// Package urlshort provides functions to implement a small server with urlshortener functionalities
package urlshort

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
			http.Redirect(w, r, redirectUrl, 301)
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

const htmlShortenResponse = `
<h2>URL Shortener</h2>
<p>Original URL: %s</p>
<p>Shortened URL: <a href="%s">%s</a></p>
<form method="post" action="/shorten">
	<input type="text" name="url" placeholder="Enter a URL">
	<input type="submit" value="Shorten">
</form>
`

// UrlShortSaver defines a contract for types that know how to save a shortened URL key.
// Types implementing this interface must provide a Save method that takes a string representing the key
// and returns an error if the operation fails.
type UrlShortSaver interface {
	// Save is a method that takes a string key representing the URL to be saved.
	// It attempts to save the key and returns an error if the operation fails.
	Save(ctx context.Context, key string, url string) error
}

// Shortener generates an HTTP handler that accepts POST requests containing a URL.
// It then generates a shortened key for the provided URL and saves it using the provided Saver.
// The generated shortened URL is displayed in the HTML response along with the original URL.
func Shortener(saver UrlShortSaver) http.HandlerFunc {
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
			http.Error(w, "invalid URL", http.StatusBadRequest)
			return
		}

		shortKey := generateShortKey()
		if err = saver.Save(r.Context(), shortKey, originalURL); err != nil {
			http.Error(w, "error saving short url", http.StatusInternalServerError)
			return
		}

		shortenedURL := fmt.Sprintf("http://%s:%s/short/%s", r.URL.Host, r.URL.Port(), shortKey)

		w.Header().Set("Content-Type", "text/html")
		responseHTML := fmt.Sprintf(htmlShortenResponse, originalURL, shortenedURL, shortenedURL)
		fmt.Fprintf(w, responseHTML)
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

// ShortenerHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys) to their corresponding URL (values
// that UrlShortGetter retrieves, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func ShortenerHandler(getter UrlShortGetter, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
		redirectUrl, err := getter.Get(r.Context(), strings.TrimRight(r.URL.Path, "/ "))
		if errors.Is(err, ErrMissingKey) {
			fallback.ServeHTTP(w, r)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, redirectUrl, 301)
	}
}
