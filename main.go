package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	r := gin.Default()

	// set up db
	db := DBSetup()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	// curl -d '{"username": "jolene"}' -H "Content-Type: application/json" -X POST localhost:8080/users
	r.POST("/users", func(c *gin.Context) {
		var json User
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var newUser UsersT
		newUser.UserName = json.Username
		//	newUser.Messages = []MessagesT{MessagesT{SentAt: time.Now()}}
		result := db.Omit("Messages").Create(&newUser)
		if result.RowsAffected == 0 {
			c.JSON(http.StatusConflict, "user with the same username already registered")
		}
		if result.Error != nil {
			// of course we wouldn't return the raw error in a prod env
			c.JSON(http.StatusInternalServerError, gin.H{"error": "problem creating user: " + result.Error.Error()})
			return
		}
		db.Save(&newUser)

		c.JSON(http.StatusOK, newUser)
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

	// curl -d '{"sender": "jolene", "recipient": {"username": "manu"}, "subject": "test subject", "body": "hello there"}' -H "Content-Type: application/json" -X POST localhost:8080/messages
	r.POST("/messages", func(c *gin.Context) {
		var cm ComposedMessage
		if err := c.ShouldBindJSON(&cm); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var m MessagesT
		m.Body = cm.Body
		m.Subject = cm.Subject
		m.Sender = cm.Sender
		if cm.Recipient.Username != "" {
			m.UserRecipient = cm.Recipient.Username
		} else {
			m.GroupRecipient = cm.Recipient.Groupname
		}

		m.SentAt = time.Now()

		db.Create(&m)

		// send message to recipient(s)
		switch {
		case m.UserRecipient != "":
			var user UsersT
			db.Where("username = ?", m.UserRecipient).Find(&user)

			// Start Association Mode
			db.Model(&user).Association("Messages")
			// `user` is the source model, it must contain primary key
			// `Messages` is a relationship's field name
			// If the above two requirements matched, the AssociationMode should be started successfully, or it should return error
			if db.Model(&user).Association("Messages").Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "problem fetching messages"})
				return
			}

			db.Model(&user).Association("Messages").Append(&m)

			// TODO: handle groups
		}

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

		var m MessagesT
		m.Body = replyMsg.Body
		// set the reply id for the message being replied to
		m.RE = mID
		m.SentAt = time.Now()

		db.Create(&m)

		// send message to sender
		var user UsersT
		db.Where("username = ?", replyMsg.Sender).Find(&user)

		// Start Association Mode
		db.Model(&user).Association("Messages")
		// `user` is the source model, it must contain primary key
		// `Messages` is a relationship's field name
		// If the above two requirements matched, the AssociationMode should be started successfully, or it should return error
		if db.Model(&user).Association("Messages").Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "problem fetching messages"})
			return
		}

		db.Model(&user).Association("Messages").Append(&m)

		c.JSON(http.StatusOK, m)
	})

	// curl localhost:8080/messages/2
	// retrieves a previously sent message
	r.GET("/messages/:id", func(c *gin.Context) {
		msgID := c.Param("id")
		mID, err := strconv.Atoi(msgID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id must be an integer: " + err.Error()})
			return
		}

		var m MessagesT
		result := db.First(&m, mID)

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("message id %d not found", mID)})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "problem fetching record"})
			return
		}

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

		var messages []MessagesT
		db.Where("re = ?", mID).Find(&messages)

		c.JSON(http.StatusOK, messages)
	})

	// retrieves a user's messages
	r.GET("/users/:username/mailbox", func(c *gin.Context) {
		username := c.Param("username")

		// Start Association Mode
		var user UsersT

		db.Where("username = ?", username).Find(&user)
		db.Model(&user).Association("Messages")
		// `user` is the source model, it must contain primary key
		// `Messages` is a relationship's field name
		// If the above two requirements matched, the AssociationMode should be started successfully, or it should return error
		if db.Model(&user).Association("Messages").Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "problem fetching messages"})
			return
		}

		var messages []MessagesT
		db.Model(&user).Association("Messages").Find(&messages)

		c.JSON(http.StatusOK, messages)
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

// Recipient is a catch-all type for the possible recipients
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
