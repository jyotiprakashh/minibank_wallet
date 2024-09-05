// deposit-service.go
package main

import (
    "encoding/json"
    "log"
    "github.com/IBM/sarama"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type Deposit struct {
    AccountID uint    `json:"account_id"`
    Amount    float64 `json:"amount"`
}

var db *gorm.DB

func initDB() {
    dsn := "host=terraform-20240902193604832600000001.cjskccquwmmt.us-east-1.rds.amazonaws.com user=jyoti password=12345789 dbname=terraform-20240902193604832600000001 port=5432 sslmode=require TimeZone=Asia/Shanghai"
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
}

func processDepositMessage(msg *sarama.ConsumerMessage) {
    var deposit Deposit
    err := json.Unmarshal(msg.Value, &deposit)
    if err != nil {
        log.Printf("Error parsing deposit message: %v", err)
        return
    }

    // Update balance in the database
    db.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", deposit.Amount, deposit.AccountID)
    log.Printf("Deposit successful for account %d: Amount %f", deposit.AccountID, deposit.Amount)
}

func main() {
    initDB()

    // Kafka Consumer Setup
    config := sarama.NewConfig()
    consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
    if err != nil {
        log.Fatalf("Error creating Kafka consumer: %v", err)
    }
    defer consumer.Close()

    // Subscribe to the "make-deposit" topic
    partitionConsumer, err := consumer.ConsumePartition("make-deposit", 0, sarama.OffsetNewest)
    if err != nil {
        log.Fatalf("Error starting consumer for partition: %v", err)
    }
    defer partitionConsumer.Close()

    // Process messages
    for msg := range partitionConsumer.Messages() {
        processDepositMessage(msg)
    }
}
