package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

var kafkaProducer sarama.SyncProducer
var responseChannels = make(map[string]chan []byte)

// Initialize Kafka Producer and Consumers
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

	topics := []string{"get-statement-response", "check-balance-response"}

	go listenForResponses(consumer, topics)
}

func listenForResponses(consumer sarama.Consumer, topics []string) {
	for _, topic := range topics {
		partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
		if err != nil {
			log.Fatalf("Error starting consumer for partition: %v", err)
		}

		responseChannels[topic] = make(chan []byte)

		go func(partitionConsumer sarama.PartitionConsumer, topic string) {
			defer partitionConsumer.Close()
			for msg := range partitionConsumer.Messages() {
				responseChannels[topic] <- msg.Value
			}
		}(partitionConsumer, topic)
	}
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


func main() {
	initKafka()

	r := gin.Default()
	r.SetTrustedProxies([]string{"localhost"})

	// Register account endpoint
	r.POST("/register", func(c *gin.Context) {
		var request map[string]interface{}
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		message, _ := json.Marshal(request)
		sendToKafka("register-account", "", message)
		c.JSON(http.StatusAccepted, gin.H{"message": "Request sent to Kafka"})
	})

	// Deposit endpoint
	r.POST("/deposit", func(c *gin.Context) {
		var request map[string]interface{}
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		message, _ := json.Marshal(request)
		sendToKafka("make-deposit", "", message) 
		c.JSON(http.StatusAccepted, gin.H{"message": "Deposit request sent to Kafka"})
	})

	// Get statement endpoint
	r.GET("/get-statement", func(c *gin.Context) {
		request := map[string]string{
			"account_id": c.Query("account_id"),
			"start_date": c.Query("start_date"),
			"end_date":   c.Query("end_date"),
		}

		message, _ := json.Marshal(request)
		sendToKafka("get-statement-request", "", message) 

		select {
		case statementData := <-responseChannels["get-statement-response"]:
			var transactions []map[string]interface{}
			json.Unmarshal(statementData, &transactions)
			c.JSON(http.StatusOK, transactions)
		case <-time.After(10 * time.Second): 
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Timeout waiting for statement response"})
		}
	})

	// Check balance endpoint
	r.POST("/check-balance", func(c *gin.Context) {
		var request map[string]interface{}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		message, _ := json.Marshal(request)
		sendToKafka("check-balance-request", "", message) 

		select {
		case balanceData := <-responseChannels["check-balance-response"]:
			var balanceResponse map[string]interface{}
			json.Unmarshal(balanceData, &balanceResponse)
			c.JSON(http.StatusOK, balanceResponse)
		case <-time.After(10 * time.Second):
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Timeout waiting for balance response"})
		}
	})

	r.POST("/login", func(c *gin.Context) {
		var request map[string]interface{}
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		requestID := "login-" + time.Now().Format("20060102150405")

		responseChannels[requestID] = make(chan []byte)

		message, _ := json.Marshal(request)
		sendToKafka("login-request", requestID, message)

		select {
		case response := <-responseChannels[requestID]:
			var responseData map[string]interface{}
			json.Unmarshal(response, &responseData)
			c.JSON(http.StatusOK, responseData)
		case <-time.After(10 * time.Second):
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Timeout waiting for response"})
		}
	})

	// Schedule Payment endpoint
	r.POST("/schedule-payment", func(c *gin.Context) {
		var request map[string]interface{}
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		message, _ := json.Marshal(request)
		sendToKafka("schedule-payment", "", message) 
		c.JSON(http.StatusAccepted, gin.H{"message": "Request sent to Kafka"})
	})

	r.Run(":3000")
}
