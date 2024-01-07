package main

import (
	"fmt"
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

	shortenerHandler := urlshort.Shortener(storage)
	retrieverHandler := urlshort.RetrieveHandler(storage, defaultMux())

	http.HandleFunc("/shorten", shortenerHandler)
	http.HandleFunc("/", retrieverHandler)
	svr := server.New(os.Getenv("PORT"))
	svr.Start()
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", fallback)
	return mux
}

const fallbackTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
 <meta charset="UTF-8">
 <title>URL Not Found</title>
 <style>
     body {
         font-family: Arial, sans-serif;
         background-color: #f5f5f5;
         padding: 20px;
         text-align: center;
     }
     h1 {
         color: #333;
         font-size: 2.5em;
     }
     p {
         color: #666;
         font-size: 1.2em;
         padding: 20px 0;
     }
 </style>
</head>
<body>
 <h1>404</h1>
 <p>The URL you entered is not available.</p>
</body>
</html>
`

func fallback(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, fallbackTemplate)
}
