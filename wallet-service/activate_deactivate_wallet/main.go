package main

import (
    "github.com/gin-gonic/gin"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "net/http"
)

type Wallet struct {
    ID        uint `json:"id" gorm:"primaryKey"`
    Active    bool `json:"active"`
}

var db *gorm.DB

func initDB() {
    dsn := "host=terraform-20240902193604832600000001.cjskccquwmmt.us-east-1.rds.amazonaws.com user=jyoti password=12345789 dbname=terraform-20240902193604832600000001 port=5432 sslmode=require TimeZone=Asia/Shanghai"    
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }

    db.AutoMigrate(&Wallet{})
}

func activateWallet(c *gin.Context) {
    var wallet Wallet
    if err := c.ShouldBindJSON(&wallet); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    db.Model(&Wallet{}).Where("id = ?", wallet.ID).Update("active", true)
    c.JSON(http.StatusOK, gin.H{"message": "Wallet activated successfully!"})
}

func deactivateWallet(c *gin.Context) {
    var wallet Wallet
    if err := c.ShouldBindJSON(&wallet); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    db.Model(&Wallet{}).Where("id = ?", wallet.ID).Update("active", false)
    c.JSON(http.StatusOK, gin.H{"message": "Wallet deactivated successfully!"})
}

func main() {
    initDB()
    r := gin.Default()
    r.POST("/activate-wallet", activateWallet)
    r.POST("/deactivate-wallet", deactivateWallet)
    r.Run(":8089")
}
