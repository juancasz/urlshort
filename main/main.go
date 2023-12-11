package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"urlshort"
)

func main() {
	yaml := flag.String("yaml", "", "path YAML file")
	json := flag.String("json", "", "path JSON file")
	flag.Parse()

	filedata, err := readFile(yaml, json)
	if err != nil {
		log.Fatal(err)
	}

	// fallback
	mux := defaultMux()

	var handler http.HandlerFunc
	if filedata.isYAML {
		handler, err = urlshort.YAMLHandler(filedata.data, mux)
		if err != nil {
			log.Fatal(err)
		}
	} else if filedata.isJSON {
		handler, err = urlshort.JSONHandler(filedata.data, mux)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", handler)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}

type fileData struct {
	data   []byte
	isJSON bool
	isYAML bool
}

func readFile(yaml, json *string) (*fileData, error) {
	var filedata fileData

	path := ""
	if len(*yaml) > 0 && len(*json) > 0 {
		return nil, fmt.Errorf("must provide json or yaml but not both at the same time")
	} else if len(*yaml) > 0 {
		path = *yaml
		if ext := filepath.Ext(path); ext != ".yml" && ext != ".yaml" {
			return nil, fmt.Errorf("expected yml or yaml file")
		}
		filedata.isYAML = true
	} else if len(*json) > 0 {
		path = *json
		if ext := filepath.Ext(path); ext != ".json" {
			return nil, fmt.Errorf("expected json file")
		}
		filedata.isJSON = true
	} else {
		return nil, fmt.Errorf("must provide a file")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	filedata.data, err = io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return &filedata, nil
}
