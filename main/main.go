package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"urlshort"
)

func main() {
	path := flag.String("path", "", "path YAML file")
	flag.Parse()

	file, err := os.Open(*path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	yamlData, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	mux := defaultMux()
	yamlHandler, err := urlshort.YAMLHandler(yamlData, mux)
	if err != nil {
		panic(err)
	}
	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", yamlHandler)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}
