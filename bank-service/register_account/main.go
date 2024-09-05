// register-account-service.go
package main

import (
	"encoding/json"
	"log"
	"github.com/IBM/sarama"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Account struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Name     string `json:"name"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"-"`
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

func processRegisterAccount(msg *sarama.ConsumerMessage) {
	var account Account
	err := json.Unmarshal(msg.Value, &account)
	if err != nil {
		log.Printf("Error parsing message: %v", err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return
	}
	account.Password = string(hashedPassword)

	db.Create(&account)
	log.Printf("Account registered successfully for email: %s", account.Email)
}

func main() {
	initDB()

	config := sarama.NewConfig()
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalf("Error creating Kafka consumer: %v", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition("register-account", 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Error starting consumer for partition: %v", err)
	}
	defer partitionConsumer.Close()

	for msg := range partitionConsumer.Messages() {
		processRegisterAccount(msg)
	}
}
