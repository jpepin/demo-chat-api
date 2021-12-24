package main

import (
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MessagesT struct {
	gorm.Model
	RE     int    `json:"re"`
	Sender string `json:"sender"`
	// example: andrew.meredith
	UserRecipient  string `json:"recipient"`
	GroupRecipient string
	Subject        string `json:"subject"`
	// example: Lunch Plans
	Body string `json:"body"`
	// example: Want to grab something around noon this Friday?
	SentAt time.Time `json:"sentat"`
	// example: 2019-09-03T17:12:42Z
	UsersTID uint
}

type GroupsT struct {
	gorm.Model
	Name string
}

type UsersT struct {
	gorm.Model
	//ID       string
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

	// Create
	// db.Create(&MessagesT{
	// 	Body:    "this is a message body",
	// 	Subject: "cats",
	// 	SentAt:  time.Now(),
	// })

	// Read
	// var msg MessagesT
	// db.First(&msg, 1)                     // find product with integer primary key
	// db.First(&msg, "subject = ?", "cats") // find product with code D42

	// fmt.Printf("This is message with subject cats: \n %+v\n", msg)

	// Update - update product's price to 200
	// db.Model(&msg).Update("Price", 200)
	// // Update - update multiple fields
	// db.Model(&msg).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
	// db.Model(&msg).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// Delete - delete product
	//	db.Delete(&msg, 1)
}
