package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", requestHandler)
	// mux.HandleFunc("/{id}", handleItem)
	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		panic(err)
	}
}

func handleItem(w http.ResponseWriter, r *http.Request) {

	id := strings.TrimPrefix(r.URL.Path, "/")
	// w.WriteHeader(307)
	w.Header().Set("StatusCode", "307")
	w.Header().Set("Location", id)

	// fmt.Printf("You requested item with ID: %s", id)

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

	body, _ := io.ReadAll(r.Body)
	fmt.Printf("body = %s\n", body)

	// encoded := base64.RawURLEncoding.EncodeToString([]byte(body))
	// fmt.Println(encoded)

	w.Write([]byte(body))

}
