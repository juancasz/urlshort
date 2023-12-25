package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"urlshort"
)

func main() {
	yaml := flag.String("yaml", "", "path YAML file")
	json := flag.String("json", "", "path JSON file")
	listenAddress := flag.String("listen", ":8080", "Listen address.")
	flag.Parse()

	filedata, err := readFile(yaml, json)
	if err != nil {
		log.Fatal(err)
	}

	// fallback
	mux := defaultMux()
	var handlerRedirect http.HandlerFunc
	if filedata.isYAML {
		handlerRedirect, err = urlshort.YAMLHandler(filedata.data, mux)
		if err != nil {
			log.Fatal(err)
		}
	} else if filedata.isJSON {
		handlerRedirect, err = urlshort.JSONHandler(filedata.data, mux)
		if err != nil {
			log.Fatal(err)
		}
	}

	router(handlerRedirect)
	startServer(*listenAddress)
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

type saver struct {
	store map[string]string
}

func (s *saver) Save(ctx context.Context, key string, url string) error {
	s.store[key] = url
	return nil
}

func router(handlerRedirect http.HandlerFunc) {
	http.HandleFunc("/", handlerRedirect)
}

func startServer(listenAddress string) {
	log.Printf("Listening at http://%s", listenAddress)

	httpServer := http.Server{
		Addr: listenAddress,
	}

	idleConnectionsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP Server Shutdown Error: %v", err)
		}
		close(idleConnectionsClosed)
	}()

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe Error: %v", err)
	}

	<-idleConnectionsClosed

	log.Printf("Bye bye")
}
