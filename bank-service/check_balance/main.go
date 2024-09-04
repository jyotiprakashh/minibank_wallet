package main

import (
    "github.com/gin-gonic/gin"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "net/http"
)

type Account struct {
    ID      uint    `json:"id"`
    Balance float64 `json:"balance"`
}

var db *gorm.DB

func initDB() {
    dsn := "host=localhost user=postgres password=yourpassword dbname=bank port=5432 sslmode=disable TimeZone=Asia/Shanghai"
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }

    db.AutoMigrate(&Account{})
}

func checkBalance(c *gin.Context) {
    var account Account
    if err := c.ShouldBindJSON(&account); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    db.First(&account, account.ID)
    c.JSON(http.StatusOK, gin.H{"balance": account.Balance})
}

func main() {
    initDB()
    r := gin.Default()
    r.POST("/check-balance", checkBalance)
    r.Run(":8084")
}
