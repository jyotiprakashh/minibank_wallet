package main

import (
    "github.com/gin-gonic/gin"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "time"
    "net/http"
)

type Payment struct {
    ID              uint      `json:"id" gorm:"primaryKey"`
    FromAccountID   uint      `json:"from_account_id"`
    ToAccountID     uint      `json:"to_account_id"`
    Amount          float64   `json:"amount"`
    ScheduledAt     time.Time `json:"scheduled_at"`
    Recurring       bool      `json:"recurring"`
    RecurrenceCycle string    `json:"recurrence_cycle"` // "monthly", "daily", "weekly"
}

var db *gorm.DB

func initDB() {
    dsn := "host=localhost user=postgres password=yourpassword dbname=bank port=5432 sslmode=disable TimeZone=Asia/Shanghai"
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }

    db.AutoMigrate(&Payment{})
}

func schedulePayment(c *gin.Context) {
    var payment Payment
    if err := c.ShouldBindJSON(&payment); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    db.Create(&payment)
    c.JSON(http.StatusOK, gin.H{"message": "Payment scheduled successfully!"})
}

func main() {
    initDB()
    r := gin.Default()
    r.POST("/schedule-payment", schedulePayment)
    r.Run(":8083")
}
