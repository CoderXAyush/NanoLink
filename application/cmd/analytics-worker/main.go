package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ClickEvent struct {
	ShortCode string `json:"short_code"`
	Timestamp int64  `json:"timestamp"`
	UserAgent string `json:"user_agent"`
}

func main() {
	log.Println("👷 Starting NanoLink Analytics Worker...")

	// 1. Connect to Mongo
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	statsColl := mongoClient.Database("nanolink").Collection("analytics")
	log.Println("✅ Connected to MongoDB!")

	// 2. Connect to Kafka Consumer
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	var consumer sarama.Consumer
	// Retry loop for Kafka connection
	for i := 0; i < 15; i++ {
		consumer, err = sarama.NewConsumer(strings.Split(kafkaBrokers, ","), config)
		if err == nil {
			break
		}
		log.Printf("⏳ Waiting for Kafka... (%d/15)", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal("❌ Kafka connection failed")
	}
	defer consumer.Close()

	// 3. Subscribe to "clicks" topic
	partitionConsumer, err := consumer.ConsumePartition("clicks", 0, sarama.OffsetNewest)
	if err != nil {
		log.Printf("⚠️ Could not start consumer (Topic might not exist yet): %v", err)
		// Ideally we wait/retry, but for demo we just block
		select {}
	}
	defer partitionConsumer.Close()

	log.Println("✅ Worker Listening for Click Events...")

	// 4. Process Messages
	for msg := range partitionConsumer.Messages() {
		var event ClickEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("❌ Bad JSON: %v", err)
			continue
		}

		log.Printf("📥 Processing Click: %s", event.ShortCode)

		// Increment Click Count in Mongo
		filter := bson.M{"short_code": event.ShortCode}
		update := bson.M{
			"$inc": bson.M{"clicks": 1},
			"$set": bson.M{"last_clicked": event.Timestamp},
		}
		opts := options.Update().SetUpsert(true)

		_, err := statsColl.UpdateOne(context.TODO(), filter, update, opts)
		if err != nil {
			log.Printf("❌ DB Write Failed: %v", err)
		} else {
			log.Println("✅ Stats Updated in Mongo")
		}
	}
}
