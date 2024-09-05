package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var kafkaProducer sarama.SyncProducer

type Payment struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	FromAccountID   uint      `json:"from_account_id"`
	ToAccountID     uint      `json:"to_account_id"`
	Amount          float64   `json:"amount"`
	ScheduledAt     time.Time `json:"scheduled_at"`
	Recurring       bool      `json:"recurring"`
	RecurrenceCycle string    `json:"recurrence_cycle"` // "monthly", "daily", "weekly"
}

func initKafka() {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Consumer.Return.Errors = true

	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %v", err)
	}
	kafkaProducer = producer

	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalf("Error creating Kafka consumer: %v", err)
	}

	go listenForPayments(consumer)
}

func listenForPayments(consumer sarama.Consumer) {
	partitionConsumer, err := consumer.ConsumePartition("schedule-payment", 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Error starting consumer for partition: %v", err)
	}
	defer partitionConsumer.Close()

	for msg := range partitionConsumer.Messages() {
		var payment Payment
		if err := json.Unmarshal(msg.Value, &payment); err != nil {
			log.Printf("Error unmarshalling payment: %v", err)
			continue
		}
		db.Create(&payment)
	}
}

func initDB() {
	dsn := "host=terraform-20240902193604832600000001.cjskccquwmmt.us-east-1.rds.amazonaws.com user=jyoti password=12345789 dbname=terraform-20240902193604832600000001 port=5432 sslmode=require TimeZone=Asia/Shanghai"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Payment{})
}

func main() {
	initDB()
	initKafka()

	r := gin.Default()

	r.POST("/schedule-payment", func(c *gin.Context) {
		var payment Payment
		if err := c.ShouldBindJSON(&payment); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		message, _ := json.Marshal(payment)
		sendToKafka("schedule-payment", "", message)

		c.JSON(http.StatusOK, gin.H{"message": "Payment scheduled successfully!"})
	})

	r.Run(":8083")
}

func sendToKafka(topic string, requestID string, message []byte) {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(requestID),
		Value: sarama.ByteEncoder(message),
	}
	_, _, err := kafkaProducer.SendMessage(msg)
	if err != nil {
		log.Fatalf("Error sending message to Kafka: %v", err)
	}
}
