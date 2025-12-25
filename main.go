package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type URLMap struct {
	sync.Mutex
	urls map[string]string
}

var store = &URLMap{urls: make(map[string]string)}

type ShortenRequest struct {
	URL string `json:"url"`
}

func main() {
	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/", redirectHandler)

	fmt.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	shortKey := fmt.Sprintf("%d", time.Now().UnixNano())
	store.Lock()
	store.urls[shortKey] = req.URL
	store.Unlock()

	fmt.Fprintf(w, "http://localhost:8080/%s\n", shortKey)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	shortKey := r.URL.Path[1:]
	store.Lock()
	url, exists := store.urls[shortKey]
	store.Unlock()

	if !exists {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}
