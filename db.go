package main

import (
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	MessagesKey = "Messages"
	UsernameKey = "user_name"
)

type MessagesT struct {
	gorm.Model
	RE     int    `json:"re"`
	Sender string `json:"sender"`
	// example: andrew.meredith
	UserRecipient  string
	GroupRecipient string
	Subject        string `json:"subject"`
	// example: Lunch Plans
	Body string `json:"body"`
	// example: Want to grab something around noon this Friday?
	SentAt time.Time `json:"sentat"`
	// example: 2019-09-03T17:12:42Z
	// Foreign Key for UsersT table
	UsersTID uint
}

type GroupsT struct {
	gorm.Model
	Name string
}

type UsersT struct {
	gorm.Model
	UserName string
	Messages []MessagesT
}

type UserGroups struct {
	gorm.Model
	UserID  int
	GroupID int
}

func DBSetup() *gorm.DB {
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	dsn := "root:my-secret-pw@tcp(127.0.0.1:3306)/app?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&UsersT{}, &MessagesT{}, &UserGroups{}, &GroupsT{})

	return db
}
