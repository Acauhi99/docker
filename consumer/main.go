package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	contentTypeJSON     = "application/json"
	maxRetries          = 10
	retryDelay          = 5 * time.Second
	connectionTimeout   = 10 * time.Second
	operationTimeout    = 5 * time.Second
	qosPrefetchCount    = 1
	healthStatusHealthy = "healthy"
)

type Event struct {
	Device    string    `json:"device" bson:"device"`
	OS        string    `json:"os" bson:"os"`
	Type      string    `json:"tipo" bson:"tipo"`
	Value     string    `json:"valor" bson:"valor"`
	IP        string    `json:"ip" bson:"ip"`
	Region    string    `json:"region" bson:"region"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

type Consumer struct {
	mongoClient *mongo.Client
	collection  *mongo.Collection
}

func main() {
	port := getEnv("CONSUMER_PORT", "")

	consumer := &Consumer{}
	if err := consumer.connectMongo(); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer consumer.mongoClient.Disconnect(context.Background())

	rabbitConn, err := connectRabbitMQ()
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitConn.Close()

	msgs, err := setupRabbitMQConsumer(rabbitConn)
	if err != nil {
		log.Fatalf("Failed to setup consumer: %v", err)
	}

	go startHealthServer(port)

	consumer.run(msgs)
}

func (c *Consumer) connectMongo() error {
	mongoURL := getMongoURL()
	var err error

	for i := range maxRetries {
		ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
		c.mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
		cancel()
		if err == nil {
			break
		}
		log.Printf("Failed to connect to MongoDB, retrying... (%d/%d)", i+1, maxRetries)
		time.Sleep(retryDelay)
	}

	if err != nil {
		return err
	}

	dbName := getEnv("MONGO_DATABASE", "")
	collName := getEnv("MONGO_COLLECTION", "")
	c.collection = c.mongoClient.Database(dbName).Collection(collName)
	return nil
}

func connectRabbitMQ() (*amqp.Connection, error) {
	rabbitURL := getRabbitMQURL()
	var conn *amqp.Connection
	var err error

	for i := range maxRetries {
		conn, err = amqp.Dial(rabbitURL)
		if err == nil {
			return conn, nil
		}
		log.Printf("Failed to connect to RabbitMQ, retrying... (%d/%d)", i+1, maxRetries)
		time.Sleep(retryDelay)
	}

	return nil, err
}

func setupRabbitMQConsumer(conn *amqp.Connection) (<-chan amqp.Delivery, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	queueName := getEnv("RABBITMQ_QUEUE", "")
	_, err = channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	if err = channel.Qos(qosPrefetchCount, 0, false); err != nil {
		return nil, err
	}

	return channel.Consume(queueName, "", false, false, false, false, nil)
}

func (c *Consumer) run(msgs <-chan amqp.Delivery) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Consumer started, waiting for messages...")

	go func() {
		for msg := range msgs {
			c.processMessage(msg)
		}
	}()

	<-quit
	log.Println("Shutting down consumer...")
}

func (c *Consumer) processMessage(msg amqp.Delivery) {
	var event Event
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		msg.Nack(false, false)
		return
	}

	event.Timestamp = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	if _, err := c.collection.InsertOne(ctx, event); err != nil {
		log.Printf("Failed to insert into MongoDB: %v", err)
		msg.Nack(false, true)
		return
	}

	log.Printf("Event processed: %s - %s", event.Device, event.Type)
	msg.Ack(false)
}

func startHealthServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Consumer health check listening on port %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Printf("Health server error: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": healthStatusHealthy})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getMongoURL() string {
	user := getEnv("MONGO_INITDB_ROOT_USERNAME", "")
	pass := getEnv("MONGO_INITDB_ROOT_PASSWORD", "")
	host := getEnv("MONGO_HOST", "")
	port := getEnv("MONGO_PORT", "")
	return "mongodb://" + user + ":" + pass + "@" + host + ":" + port
}

func getRabbitMQURL() string {
	user := getEnv("RABBITMQ_DEFAULT_USER", "")
	pass := getEnv("RABBITMQ_DEFAULT_PASS", "")
	host := getEnv("RABBITMQ_HOST", "")
	port := getEnv("RABBITMQ_PORT", "")
	return "amqp://" + user + ":" + pass + "@" + host + ":" + port + "/"
}
