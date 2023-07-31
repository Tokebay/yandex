package main

import (
	"math/rand"
	"net/http"
	"strings"
	"time"

	bs "github.com/catinello/base62"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", requestHandler)

	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		panic(err)
	}
}

func handleItem(w http.ResponseWriter, r *http.Request) {

	id := strings.TrimPrefix(r.URL.Path, "/")

	w.Header().Set("Location", id)
	w.WriteHeader(http.StatusTemporaryRedirect)
	// fmt.Fprintf(w, "You requested item with ID: %s", id)

}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleItem(w, r)
	case "POST":
		postReqHandler(w, r)
	default:
		w.WriteHeader(400)
	}
}

func postReqHandler(w http.ResponseWriter, r *http.Request) {

	r.Header.Add("Content-type", "text/plain")
	w.WriteHeader(201)
	w.Header().Set("Content-type", "text/plain")

	rand.Seed(time.Now().UnixNano())

	encode := (bs.Encode(rand.Intn(99999999999)))
	w.Write([]byte("http://localhost:8080/" + encode))

}
