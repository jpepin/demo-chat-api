package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

func CreateMessageForUser(recipient string, db *gorm.DB, m *MessagesT, c *gin.Context) {
	var user UsersT
	db.Where("user_name = ?", recipient).Find(&user)

	// Start Association Mode
	db.Model(&user).Association(MessagesKey)
	// `user` is the source model, it must contain the primary key
	// `Messages` is a relationship's field name
	// If the above two requirements matched, the AssociationMode should be started successfully, or it should return an error
	if assocErr := db.Model(&user).Association(MessagesKey).Error; assocErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("problem with association: %s", assocErr.Error())})
		return
	}
	// add message associated with the user
	db.Model(&user).Association(MessagesKey).Append(m)
	db.Save(&user)
	c.JSON(http.StatusOK, m)
}

// GetMessage retrieves a message by id
func GetMessage(messageID int, db *gorm.DB, c *gin.Context) (MessagesT, error) {
	var m MessagesT
	result := db.First(&m, messageID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("message id %d not found", messageID)})
			return m, result.Error
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("problem fetching record: %s", result.Error.Error())})
		return m, result.Error
	}
	return m, nil
}
