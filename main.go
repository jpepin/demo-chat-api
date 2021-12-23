package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	// curl -d '{"username": "manu", "password": "123"}' -H "Content-Type: application/json" -X POST localhost:8080/users
	r.POST("/users", func(c *gin.Context) {
		// Example for binding JSON ({"user": "manu", "password": "123"})
		var json User
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if json.Username != "manu" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "user created"})
	})
	// curl -d '{"usernames": ["manu"], "groupname": "group1"}' -H "Content-Type: application/json" -X POST localhost:8080/groups
	r.POST("/groups", func(c *gin.Context) {
		var json GroupCreation
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if json.GroupName != "group1" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "group successfully created"})
	})
	r.GET("/users/:username/mailbox", func(c *gin.Context) {
		username := c.Param("username")
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("fake messages for %s", username),
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

type Message struct {
	id     int
	re     int
	sender string
	// example: andrew.meredith
	recipient Recipient
	subject   string
	// example: Lunch Plans
	body string
	// example: Want to grab something around noon this Friday?
	sentAt string
	// example: 2019-09-03T17:12:42Z
}

type Recipient interface {
	name() string
}

type User struct {
	Username string `json:"username"`
}

type GroupCreation struct {
	GroupName string   `json:"groupname"`
	Usernames []string `json:"usernames"`
}
