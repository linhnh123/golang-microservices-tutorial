package service

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func StartWebServer(port string) {
	r := NewRouter()
	http.Handle("/", r)

	logrus.Infoln("Starting HTTP service at " + port)
	err := http.ListenAndServe(":"+port, nil)

	if err != nil {
		logrus.Infoln("Error HTTP " + port)
		logrus.Infoln("Error: " + err.Error())
	}
}
