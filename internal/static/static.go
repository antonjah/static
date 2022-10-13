package static

import (
	"fmt"
	"log"
	"net/http"
)

var (
	host         = getEnv("HOST", "")
	port         = getEnv("PORT", "8080")
	responseBody = getEnv("RESPONSE", "")
	contentType  = getEnv("CONTENT_TYPE", "plain/text")
	bindAddress  = fmt.Sprintf("%s:%s", host, port)
)

func Run() {
	log.Println("static is listening on", bindAddress)
	if err := http.ListenAndServe(bindAddress, handler()); err != nil {
		log.Fatal(err)
	}
}
