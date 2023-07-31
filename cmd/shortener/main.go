package main

import (
	"math/rand"
	"net/http"
	"time"

	bs "github.com/catinello/base62"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", getReqHandler)

	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		panic(err)
	}
}

func getReqHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.WriteHeader(307)
		w.Header().Set("Location", "https://practicum.yandex.ru/")

	case "POST":
		r.Header.Add("Content-type", "text/plain")
		w.WriteHeader(201)
		w.Header().Set("Content-type", "text/plain")

		rand.Seed(time.Now().UnixNano())

		encode := (bs.Encode(rand.Intn(9999999999)))
		w.Write([]byte("http://localhost:8080/" + encode))
	default:
		w.WriteHeader(400)

	}
}
