package aggregator

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
)

var client = &http.Client{}
var logglyBaseUrl = "https://logs-01.loggly.com/inputs/%s/tag/http/"
var url string

func Start(bulkQueue chan []byte, authToken string) {
	url = fmt.Sprintf(logglyBaseUrl, authToken)
	buf := new(bytes.Buffer)
	for {
		msg := <-bulkQueue // Blocks here until a message arrives on the channel.
		buf.Write(msg)
		buf.WriteString("\n")

		size := buf.Len()
		if size > 1024 {
			sendBulk(*buf)
			buf.Reset()
		}
	}
}

func sendBulk(buffer bytes.Buffer) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(buffer.Bytes()))
	if err != nil {
		log.Println("Error creating bulk upload HTTP request: " + err.Error())
		return
	}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		log.Println("Error sending bulk: " + err.Error())
		return
	}
	log.Printf("Successfully sent batch of %v bytes to Loggly\n", buffer.Len())
}
