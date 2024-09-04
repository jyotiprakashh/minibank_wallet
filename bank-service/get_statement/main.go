package main

import (
    "github.com/gin-gonic/gin"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "time"
    "net/http"
)

type Transaction struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    AccountID   uint      `json:"account_id"`
    Amount      float64   `json:"amount"`
    CreatedAt   time.Time `json:"created_at"`
    Description string    `json:"description"`
}

var db *gorm.DB

func initDB() {
    dsn := "host=localhost user=postgres password=yourpassword dbname=bank port=5432 sslmode=disable TimeZone=Asia/Shanghai"
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }

    db.AutoMigrate(&Transaction{})
}

func getStatement(c *gin.Context) {
    var transactions []Transaction
    accountID := c.Query("account_id")
    startDate := c.Query("start_date")
    endDate := c.Query("end_date")

    db.Where("account_id = ? AND created_at BETWEEN ? AND ?", accountID, startDate, endDate).Find(&transactions)
    c.JSON(http.StatusOK, transactions)
}

func main() {
    initDB()
    r := gin.Default()
    r.GET("/get-statement", getStatement)
    r.Run(":8085")
}
