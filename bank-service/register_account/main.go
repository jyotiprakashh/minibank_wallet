package main

import (
    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "net/http"
)

type Account struct {
    ID       uint   `json:"id" gorm:"primaryKey"`
    Name     string `json:"name"`
    Email    string `json:"email" gorm:"unique"`
    Password string `json:"-"`
}

var db *gorm.DB

func initDB() {
    dsn := "host=localhost user=postgres password=password dbname=bank port=5432 sslmode=disable TimeZone=Asia/Shanghai"
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }

    db.AutoMigrate(&Account{})
}

func registerAccount(c *gin.Context) {
    var account Account
    if err := c.ShouldBindJSON(&account); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }
    account.Password = string(hashedPassword)

    db.Create(&account)
    c.JSON(http.StatusOK, gin.H{"message": "Account registered successfully!"})
}

func main() {
    initDB()
    r := gin.Default()
    r.POST("/register", registerAccount)
    r.Run(":8080")
}
