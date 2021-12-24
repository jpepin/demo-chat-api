package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	MessagesKey = "Messages"
	UsernameKey = "user_name"
)

type MessagesT struct {
	gorm.Model
	// The message being responded to, if any
	RE int `json:"re"`
	// example: andrew.meredith
	Sender        string `json:"sender"`
	UserRecipient string
	// only set if this was part of a group message
	GroupRecipient string
	// example: Lunch Plans
	Subject string `json:"subject"`
	// example: Want to grab something around noon this Friday?
	Body string `json:"body"`
	// example: 2019-09-03T17:12:42Z
	SentAt time.Time `json:"sentat"`
	// Foreign Key for UsersT table
	UsersTID uint
}

func (m MessagesT) Send(db *gorm.DB, c *gin.Context) {
	// send message to recipient(s)
	switch {
	case m.UserRecipient != "":
		CreateMessageForUser(m.UserRecipient, db, &m, c)
		// TODO: handle groups
	}
}

func (m MessagesT) Reply(db *gorm.DB, c *gin.Context) {
	// find sender of original message
	originalMessage, err := GetMessage(m.RE, db, c)
	if err != nil {
		return
	}
	replyTo := make(map[string]bool)

	// check if it was a group message
	if originalMessage.GroupRecipient != "" {
		// TODO look up group members and add
		m.GroupRecipient = originalMessage.GroupRecipient
	}
	replyTo[originalMessage.Sender] = true

	// send message back to sender(s) and/or group
	for replyToUser := range replyTo {
		m.UserRecipient = replyToUser
		CreateMessageForUser(replyToUser, db, &m, c)
	}
}

func FromComposedMessage(cm ComposedMessage) MessagesT {
	var m MessagesT
	m.Body = cm.Body
	m.Subject = cm.Subject
	m.Sender = cm.Sender
	m.UserRecipient = cm.Recipient.Username
	m.GroupRecipient = cm.Recipient.Groupname
	m.SentAt = time.Now()

	return m
}

func FromReplyMessage(response ReplyMessage, originalMessageID int) MessagesT {
	var m MessagesT
	m.Body = response.Body
	// set the reply id for the message being replied to
	m.RE = originalMessageID
	m.SentAt = time.Now()
	m.Sender = response.Sender

	return m
}

type GroupsT struct {
	gorm.Model
	Name string `gorm:"index:idx_name,unique"`
}

// TODO: use gorm many-to-many association here
type UserGroup struct {
	gorm.Model
	GroupName string
	Username  string
}

type Usernames string

type UsersT struct {
	gorm.Model
	UserName string
	Messages []MessagesT
}

func DBSetup() *gorm.DB {
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	dsn := "root:my-secret-pw@tcp(127.0.0.1:3306)/app?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&UsersT{}, &MessagesT{}, &GroupsT{}, &UserGroup{})

	return db
}
