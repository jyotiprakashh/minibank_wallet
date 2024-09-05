// balance-service.go
package main

import (
    "encoding/json"
    "log"

    "github.com/IBM/sarama"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type Account struct {
    ID      uint    `json:"id"`
    Balance float64 `json:"balance"`
}

var db *gorm.DB

func initDB() {
    dsn := "host=terraform-20240902193604832600000001.cjskccquwmmt.us-east-1.rds.amazonaws.com user=jyoti password=12345789 dbname=terraform-20240902193604832600000001 port=5432 sslmode=require TimeZone=Asia/Shanghai"
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }

    db.AutoMigrate(&Account{})
}

func processBalanceRequest(msg *sarama.ConsumerMessage, producer sarama.SyncProducer) {
    var request map[string]interface{}
    err := json.Unmarshal(msg.Value, &request)
    if err != nil {
        log.Printf("Error parsing request: %v", err)
        return
    }

    var account Account
    db.First(&account, request["id"])

    response := map[string]interface{}{
        "id":      account.ID,
        "balance": account.Balance,
    }

    responseMessage, _ := json.Marshal(response)
    sendToKafka("check-balance-response", responseMessage, producer)
}

func sendToKafka(topic string, message []byte, producer sarama.SyncProducer) {
    msg := &sarama.ProducerMessage{
        Topic: topic,
        Value: sarama.ByteEncoder(message),
    }
    _, _, err := producer.SendMessage(msg)
    if err != nil {
        log.Fatalf("Error sending message to Kafka: %v", err)
    }
}


func main() {
    initDB()

    config := sarama.NewConfig()
    consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
    if err != nil {
        log.Fatalf("Error creating Kafka consumer: %v", err)
    }
    defer consumer.Close()

    partitionConsumer, err := consumer.ConsumePartition("check-balance-request", 0, sarama.OffsetNewest)
    if err != nil {
        log.Fatalf("Error starting consumer for partition: %v", err)
    }
    defer partitionConsumer.Close()

    producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
    if err != nil {
        log.Fatalf("Error creating Kafka producer: %v", err)
    }
    defer producer.Close()

    for msg := range partitionConsumer.Messages() {
        processBalanceRequest(msg, producer)
    }
}
