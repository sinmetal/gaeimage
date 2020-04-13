package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	v1 "github.com/sinmetal/gaeimage"
	v2 "github.com/sinmetal/gaeimage/v2"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	http.HandleFunc("/v2/", v2.ImageHandler)
	http.HandleFunc("/v1/", v1.ImageHandler)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), http.DefaultServeMux); err != nil {
		log.Printf("failed ListenAndServe err=%+v", err)
	}
}
