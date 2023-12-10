package urlshort

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
		if redirectUrl, ok := pathsToUrls[strings.TrimRight(r.URL.Path, "/ ")]; ok {
			http.Redirect(w, r, redirectUrl, 301)
			return
		}
		fallback.ServeHTTP(w, r)
	}
}

type URLMapper struct {
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
	var mappers []URLMapper
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

func JSONHandler(data []byte, fallback http.Handler) (http.HandlerFunc, error) {
	var mappers []URLMapper
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

func buildMap(mappers []URLMapper) (map[string]string, error) {
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
