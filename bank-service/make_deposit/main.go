package main

import (
    "github.com/gin-gonic/gin"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "net/http"
)

type Deposit struct {
    AccountID uint    `json:"account_id"`
    Amount    float64 `json:"amount"`
}

var db *gorm.DB

func initDB() {
    dsn := "host=localhost user=postgres password=yourpassword dbname=bank port=5432 sslmode=disable TimeZone=Asia/Shanghai"
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
}

func makeDeposit(c *gin.Context) {
    var deposit Deposit
    if err := c.ShouldBindJSON(&deposit); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Update balance in the database
    db.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", deposit.Amount, deposit.AccountID)
    c.JSON(http.StatusOK, gin.H{"message": "Deposit successful"})
}

func main() {
    initDB()
    r := gin.Default()
    r.POST("/deposit", makeDeposit)
    r.Run(":8082")
}
