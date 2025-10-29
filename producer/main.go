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
)

const (
	contentTypeJSON   = "application/json"
	maxRetries        = 10
	retryDelay        = 5 * time.Second
	operationTimeout  = 5 * time.Second
	serverReadTimeout = 10 * time.Second
	serverWriteTimeout = 10 * time.Second
	shutdownTimeout   = 5 * time.Second
	statusAccepted    = "accepted"
	statusHealthy     = "healthy"
)

type Event struct {
	Device string `json:"device"`
	OS     string `json:"os"`
	Type   string `json:"tipo"`
	Value  string `json:"valor"`
	IP     string `json:"ip"`
	Region string `json:"region"`
}

type Producer struct {
	rabbitConn    *amqp.Connection
	rabbitChannel *amqp.Channel
	queueName     string
}

func main() {
	port := getEnv("PRODUCER_PORT", "")

	producer := &Producer{}
	if err := producer.connectRabbitMQ(); err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer producer.rabbitConn.Close()
	defer producer.rabbitChannel.Close()

	server := producer.startServer(port)
	producer.waitForShutdown(server)
}

func (p *Producer) connectRabbitMQ() error {
	rabbitURL := getRabbitMQURL()
	var err error

	for i := range maxRetries {
		p.rabbitConn, err = amqp.Dial(rabbitURL)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ, retrying... (%d/%d)", i+1, maxRetries)
		time.Sleep(retryDelay)
	}

	if err != nil {
		return err
	}

	p.rabbitChannel, err = p.rabbitConn.Channel()
	if err != nil {
		return err
	}

	p.queueName = getEnv("RABBITMQ_QUEUE", "")
	_, err = p.rabbitChannel.QueueDeclare(p.queueName, true, false, false, false, nil)
	return err
}

func (p *Producer) startServer(port string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/events", p.handleEvents)
	mux.HandleFunc("/health", handleHealth)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  serverReadTimeout,
		WriteTimeout: serverWriteTimeout,
	}

	go func() {
		log.Printf("Producer API listening on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	return server
}

func (p *Producer) waitForShutdown(server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}

func (p *Producer) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if !p.validateEvent(event) {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	if err := p.publishEvent(event); err != nil {
		log.Printf("Failed to publish message: %v", err)
		http.Error(w, "Failed to process event", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]string{"status": statusAccepted})
}

func (p *Producer) validateEvent(event Event) bool {
	return event.Device != "" && event.OS != "" && event.Type != ""
}

func (p *Producer) publishEvent(event Event) error {
	body, _ := json.Marshal(event)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	return p.rabbitChannel.PublishWithContext(
		ctx,
		"",
		p.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  contentTypeJSON,
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": statusHealthy})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getRabbitMQURL() string {
	user := getEnv("RABBITMQ_DEFAULT_USER", "")
	pass := getEnv("RABBITMQ_DEFAULT_PASS", "")
	host := getEnv("RABBITMQ_HOST", "")
	port := getEnv("RABBITMQ_PORT", "")
	return "amqp://" + user + ":" + pass + "@" + host + ":" + port + "/"
}
