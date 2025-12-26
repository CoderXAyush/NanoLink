package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/CoderXAyush/NanoLink/internal/base62"
	"github.com/CoderXAyush/NanoLink/internal/store"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	redisClient   *redis.Client
	mongoClient   *mongo.Client
	idGen         *store.IDGenerator
	collection    *mongo.Collection
	kafkaProducer sarama.SyncProducer
)

type ShortenRequest struct {
	LongURL string `json:"long_url"`
}

type ShortenResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
}

type LinkDoc struct {
	ID        uint64 `bson:"_id"`
	ShortCode string `bson:"short_code"`
	LongURL   string `bson:"long_url"`
	CreatedAt int64  `bson:"created_at"`
}

type ClickEvent struct {
	ShortCode string `json:"short_code"`
	Timestamp int64  `json:"timestamp"`
	UserAgent string `json:"user_agent"`
}

func main() {
	log.Println("🚀 Starting NanoLink API Server...")

	// 1. Connect Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient = redis.NewClient(&redis.Options{Addr: redisAddr})

	// 2. Connect Mongo
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoClient, _ = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	collection = mongoClient.Database("nanolink").Collection("links")

	// 3. Connect Kafka Producer
	kafkaAddr := os.Getenv("KAFKA_BROKERS")
	if kafkaAddr == "" {
		kafkaAddr = "localhost:9092"
	}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	var err error
	kafkaProducer, err = sarama.NewSyncProducer([]string{kafkaAddr}, config)
	if err != nil {
		log.Printf("⚠️ Kafka Producer failed: %v", err)
	} else {
		log.Println("✅ Kafka Producer Connected!")
	}

	// 4. Init ID Generator
	idGen = store.NewIDGenerator(redisClient)

	// 5. Routes
	http.HandleFunc("/api/v1/shorten", handleShorten)
	http.HandleFunc("/", handleRedirect)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🌍 Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	id, err := idGen.GetID(r.Context())
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}
	shortCode := base62.Encode(id)

	doc := LinkDoc{ID: id, ShortCode: shortCode, LongURL: req.LongURL, CreatedAt: time.Now().Unix()}
	collection.InsertOne(r.Context(), doc)
	redisClient.Set(r.Context(), shortCode, req.LongURL, 24*time.Hour)

	resp := ShortenResponse{ShortCode: shortCode, ShortURL: fmt.Sprintf("http://%s/%s", r.Host, shortCode)}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	shortCode := r.URL.Path[1:]
	if shortCode == "" {
		return
	}

	// 1. Resolve URL (Redis -> Mongo)
	longURL, err := redisClient.Get(r.Context(), shortCode).Result()
	if err != nil {
		var doc LinkDoc
		if err := collection.FindOne(r.Context(), bson.M{"short_code": shortCode}).Decode(&doc); err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		longURL = doc.LongURL
		redisClient.Set(r.Context(), shortCode, longURL, 24*time.Hour)
	}

	// 2. Fire Async Analytics Event (Don't wait for it)
	go func() {
		if kafkaProducer == nil {
			return
		}
		event := ClickEvent{ShortCode: shortCode, Timestamp: time.Now().Unix(), UserAgent: r.UserAgent()}
		bytes, _ := json.Marshal(event)

		_, _, err := kafkaProducer.SendMessage(&sarama.ProducerMessage{
			Topic: "clicks",
			Value: sarama.ByteEncoder(bytes),
		})
		if err != nil {
			log.Printf("❌ Failed to send Kafka event: %v", err)
		} else {
			log.Printf("📤 Sent Click Event: %s", shortCode)
		}
	}()

	// 3. Redirect User
	http.Redirect(w, r, longURL, http.StatusFound)
}
