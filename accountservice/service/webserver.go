package service

import (
	"log"
	"net/http"
)

func StartWebServer(port string) {
	r := NewRouter()
	http.Handle("/", r)

	log.Println("Starting HTTP service at " + port)
	err := http.ListenAndServe(":"+port, nil)

	if err != nil {
		log.Println("Error HTTP " + port)
		log.Println("Error: " + err.Error())
	}
}
