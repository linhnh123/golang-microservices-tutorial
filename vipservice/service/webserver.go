package service

import (
	"fmt"
	"net/http"
)

func StartWebServer(port string) {
	r := NewRouter()
	http.Handle("/", r)

	fmt.Println("Starting HTTP service at " + port)
	err := http.ListenAndServe(":"+port, nil)

	if err != nil {
		fmt.Println("An error occured starting HTTP listener at port " + port)
		fmt.Println("Error: " + err.Error())
	}
}
