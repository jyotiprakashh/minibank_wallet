package main

import (
    "github.com/dgrijalva/jwt-go"
    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "time"
    "net/http"
)

var jwtKey = []byte("your_secret_key")
var db *gorm.DB

type Account struct {
    ID       uint   `json:"id"`
    Email    string `json:"email"`
    Password string `json:"-"`
}

func initDB() {
    dsn := "host=localhost user=postgres password=password dbname=bank port=5432 sslmode=disable TimeZone=Asia/Shanghai"
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
}

func login(c *gin.Context) {
    var account Account
    var input Account
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    db.Where("email = ?", input.Email).First(&account)

    if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(input.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "id":  account.ID,
        "exp": time.Now().Add(time.Hour * 24).Unix(),
    })
    tokenString, _ := token.SignedString(jwtKey)

    c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func main() {
    initDB()
    r := gin.Default()
    r.POST("/login", login)
    r.Run(":8081")
}
