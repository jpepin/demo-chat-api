package main

import (
	"fmt"
	"net/http"
	"strconv"

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

		c.JSON(http.StatusOK, json)
	})

	// curl -d '{"sender": "Jolene", "recipient": {"username": "manu"}, "subject": "test subject", "body": "hello there"}' -H "Content-Type: application/json" -X POST localhost:8080/messages
	r.POST("/messages", func(c *gin.Context) {
		var cm ComposedMessage
		if err := c.ShouldBindJSON(&cm); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var m Message
		m.Body = cm.Body
		m.Recipient.Username = cm.Recipient.Username
		c.JSON(http.StatusOK, m)
	})

	// curl -d '{"sender": "Jolene", "recipient": {"username": "manu"}, "subject": "test subject", "body": "hello there"}' -H "Content-Type: application/json" -X POST localhost:8080/messages/2/replies
	r.POST("/messages/:id/replies", func(c *gin.Context) {
		var replyMsg ReplyMessage
		msgID := c.Param("id")
		mID, err := strconv.Atoi(msgID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := c.ShouldBindJSON(&replyMsg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var m Message
		m.Body = replyMsg.Body
		m.ID = mID
		c.JSON(http.StatusOK, m)
	})

	// curl localhost:8080/messages/2
	r.GET("/messages/:id", func(c *gin.Context) {
		msgID := c.Param("id")
		if msgID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing message id"})
			return
		}

		var m Message
		m.Body = "fake body"
		m.Recipient.Username = "someone"
		c.JSON(http.StatusOK, m)
	})

	// curl localhost:8080/messages/2/replies
	r.GET("/messages/:id/replies", func(c *gin.Context) {
		msgID := c.Param("id")
		mID, err := strconv.Atoi(msgID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var m Message
		m.Body = "fake body"
		m.Recipient.Username = "someone"
		m.ID = mID
		msgs := []Message{m}
		c.JSON(http.StatusOK, msgs)
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
	ID     int    `json:"id"`
	RE     int    `json:"re"`
	Sender string `json:"sender"`
	// example: andrew.meredith
	Recipient Recipient `json:"recipient"`
	Subject   string    `json:"subject"`
	// example: Lunch Plans
	Body string `json:"body"`
	// example: Want to grab something around noon this Friday?
	SentAt string `json:"sentat"`
	// example: 2019-09-03T17:12:42Z
}

// Recipient is a catch-all type for the posisble recipients
// because I don't want to implement custom marshaling rn
type Recipient struct {
	Username  string `json:"username"`
	Groupname string `json:"groupname"`
}

// type UserRecipient struct {
// 	// description:	A message recipient representing a single user
// 	username string
// 	// example: andrew.meredith
// }

// type GroupRecipient struct {
// 	// description: A message recipient representing a group of users
// 	groupname string
// 	// example: quantummetric
// }

type User struct {
	Username string `json:"username"`
}

type GroupCreation struct {
	GroupName string   `json:"groupname"`
	Usernames []string `json:"usernames"`
}

type ComposedMessage struct {
	Sender string `json:"sender"`
	// example: andrew.meredith
	Recipient Recipient `json:"recipient"`
	Subject   string    `json:"subject"`
	// example: Lunch Plans
	Body string `json:"body"`
	// example: Want to grab something around noon this Friday?
}

type ReplyMessage struct {
	Sender string `json:"sender"`
	// example: andrew.meredith
	Subject string `json:"subject"`
	// example: Lunch Plans
	Body string `json:"body"`
	// example: Want to grab something around noon this Friday?
}
