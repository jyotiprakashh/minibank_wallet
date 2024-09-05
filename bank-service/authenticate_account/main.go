// auth-microservice.go
package main

import (
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var kafkaProducer sarama.SyncProducer

type Account struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

// Initialize Kafka Producer
func initKafkaProducer() {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %v", err)
	}
	kafkaProducer = producer
}

// Initialize the database
func initDB() {
	dsn := "host=terraform-20240902193604832600000001.cjskccquwmmt.us-east-1.rds.amazonaws.com user=jyoti password=12345789 dbname=terraform-20240902193604832600000001 port=5432 sslmode=require TimeZone=Asia/Shanghai"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
}

func main() {
	initDB()
	initKafkaProducer()

	config := sarama.NewConfig()
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalf("Error creating Kafka consumer: %v", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition("login-request", 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Error starting consumer for partition: %v", err)
	}
	defer partitionConsumer.Close()

	for msg := range partitionConsumer.Messages() {
		requestID := string(msg.Key)
		var request map[string]interface{}
		json.Unmarshal(msg.Value, &request)

		var account Account
		db.Where("email = ?", request["email"]).First(&account)

		response := make(map[string]interface{})
		if bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(request["password"].(string))) == nil {
			response["status"] = "success"
			response["account_id"] = account.ID
		} else {
			response["status"] = "failed"
			response["message"] = "Invalid credentials"
		}

		responseMessage, _ := json.Marshal(response)
		sendToKafka("service-response", requestID, responseMessage)
	}
}

// Send the response to Kafka
func sendToKafka(topic string, key string, message []byte) {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(message),
	}
	_, _, err := kafkaProducer.SendMessage(msg)
	if err != nil {
		log.Fatalf("Error sending message to Kafka: %v", err)
	}
}
