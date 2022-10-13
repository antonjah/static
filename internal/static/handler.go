package static

import (
	"log"
	"net/http"
)

func handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", contentType)
		if _, err := w.Write([]byte(responseBody)); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
