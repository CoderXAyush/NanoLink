package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

type URLDocument struct {
	ID       string `bson:"_id"`
	Original string `bson:"original_url"`
}

func main() {
	// 1. Connect to MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017" // Fallback for local testing
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("Could not connect to MongoDB:", err)
	}
	fmt.Println("Connected to MongoDB!")

	collection = client.Database("shortener").Collection("urls")

	// 2. Define Routes
	http.HandleFunc("/api/shorten", shortenHandler) // POST to shorten
	http.HandleFunc("/", redirectHandler)           // GET to redirect

	// 3. Start Server
	port := ":8080"
	log.Printf("Server listening on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// Handler to Shorten URL
func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate a random 6-character ID
	shortID := generateShortID()

	// Save to MongoDB
	doc := URLDocument{ID: shortID, Original: req.URL}
	_, err := collection.InsertOne(context.TODO(), doc)
	if err != nil {
		http.Error(w, "Failed to save URL", http.StatusInternalServerError)
		return
	}

	// Return the result
	resp := ShortenResponse{
		ShortURL: fmt.Sprintf("http://%s/%s", r.Host, shortID),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Handler to Redirect
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	// Remove the leading slash to get the ID
	shortID := r.URL.Path[1:]

	if shortID == "" {
		w.Write([]byte("Distributed URL Shortener API is running"))
		return
	}

	// Find in MongoDB
	var result URLDocument
	err := collection.FindOne(context.TODO(), bson.M{"_id": shortID}).Decode(&result)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	// Redirect!
	http.Redirect(w, r, result.Original, http.StatusFound)
}

func generateShortID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
