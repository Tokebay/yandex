// package main

// import (
// 	"fmt"
// 	"io/ioutil"
// 	"math/rand"
// 	"net/http"
// 	"net/url"
// 	"strings"
// 	"time"
// )

// var shortenedURLs = make(map[string]string)

// func main() {
// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/", requestHandler)

// 	err := http.ListenAndServe(":8080", mux)

// 	if err != nil {
// 		panic(err)
// 	}
// }

// func handleItem(w http.ResponseWriter, r *http.Request) {

// 	id := strings.TrimPrefix(r.URL.Path, "/")
// 	//id := r.URL.Path[1:]

// 	decodedURL, err := url.QueryUnescape(id)
// 	if err != nil {
// 		http.Error(w, "Error decoding URL", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Location", decodedURL)
// 	fmt.Printf("You requested item with ID: %s", decodedURL)
// 	w.WriteHeader(http.StatusTemporaryRedirect)

// 	// fmt.Printf("You requested item with ID: %s", id)

// }

// func requestHandler(w http.ResponseWriter, r *http.Request) {
// 	switch r.Method {
// 	case "GET":
// 		handleItem(w, r)
// 	case "POST":
// 		postReqHandler(w, r)
// 	default:
// 		w.WriteHeader(400)
// 	}
// }

// func postReqHandler(w http.ResponseWriter, r *http.Request) {

// 	w.Header().Set("Content-type", "text/plain")

// 	body, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, "Error reading request body", http.StatusBadRequest)
// 		return
// 	}
// 	defer r.Body.Close()
// 	// shortenedURL := base62Encode(body)
// 	shortenedURL := shortenURL(string(body))

// 	// Отвечаем клиенту с сокращённым URL и кодом 201
// 	w.Header().Set("Content-Type", "text/plain")
// 	w.WriteHeader(http.StatusCreated)
// 	fmt.Fprintf(w, shortenedURL)

// }

// const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// func generateRandomString(length int) string {
// 	b := make([]byte, length)
// 	for i := range b {
// 		b[i] = charset[seededRand.Intn(len(charset))]
// 	}
// 	return string(b)
// }

// func shortenURL(originalURL string) string {

// 	shortCode := generateRandomString(8) // Можно задать другую длину для сокращённого URL
// 	baseURL := "http://localhost:8080/"  // Замените на ваш домен
// 	shortenedURL := baseURL + shortCode

// 	return shortenedURL
// }

// const base62Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// func base62Encode(num int) string {
// 	// Адаптируем Base64 encoding для Base62 алфавита
// 	base := len(base62Alphabet)
// 	var encoded []byte

// 	for num > 0 {
// 		remainder := num % base
// 		num = num / base
// 		encoded = append([]byte{base62Alphabet[remainder]}, encoded...)
// 	}

// 	if len(encoded) == 0 {
// 		return string(base62Alphabet[0])
// 	}

// 	return string(encoded)
// }
