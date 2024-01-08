package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"urlshort"
	"urlshort/internal/redis"
	"urlshort/internal/server"
)

func main() {
	expirationMinutes, err := strconv.Atoi(os.Getenv("REDIS_EXPIRATION_MINUTES"))
	if err != nil {
		log.Fatal(err)
	}
	storage := redis.New(&redis.Options{
		Host:              os.Getenv("REDIS_HOST"),
		Port:              os.Getenv("REDIS_PORT"),
		Username:          os.Getenv("REDIS_USERNAME"),
		Password:          os.Getenv("REDIS_PASSWORD"),
		ExpirationMinutes: expirationMinutes,
	})

	shortenerHandler := urlshort.Shortener(storage, os.Getenv("HOST"), invalidUrlMux())
	retrieverHandler := urlshort.RetrieveHandler(storage, missingUrlMux())

	http.HandleFunc("/home", urlshort.ShortenerHome)
	http.HandleFunc("/shorten", shortenerHandler)
	http.HandleFunc("/short/", retrieverHandler)
	svr := server.New(os.Getenv("PORT"))
	svr.Start()
}

func missingUrlMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", urlshort.MissingUrlHandler)
	return mux
}

func invalidUrlMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", urlshort.InvalidUrlHandler)
	return mux
}
