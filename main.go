package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

func main() {

	// set up and migrate db
	db := DBSetup()

	r := setupRouter()
	// curl -d '{"username": "jolene"}' -H "Content-Type: application/json" -X POST localhost:8080/users
	r.POST("/users", func(c *gin.Context) {
		var json User
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var newUser UsersT
		newUser.UserName = json.Username
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

		newGroup := GroupsT{
			Name: json.GroupName,
		}

		result := db.Create(&newGroup)
		if result.Error != nil {
			// Check if duplicate entry
			mysqlErr := result.Error.(*mysql.MySQLError)
			switch mysqlErr.Number {
			case 1062:
				c.JSON(http.StatusConflict, "group with the same name already registered")
				return
			}
			// of course we wouldn't return the raw error in a prod env
			c.JSON(http.StatusInternalServerError, gin.H{"error": "problem creating group: " + result.Error.Error()})
			return
		}
		// add users to group
		for _, username := range json.Usernames {
			userGroupMembership := UserGroup{
				GroupName: json.GroupName,
				Username:  username,
			}
			userResult := db.Create(&userGroupMembership)
			if userResult.Error != nil {
				// of course we wouldn't return the raw error in a prod env
				c.JSON(http.StatusInternalServerError, gin.H{"error": "problem adding user to group: " + result.Error.Error()})
				return
			}
		}

		c.JSON(http.StatusCreated, json)
	})

	// curl -d '{"sender": "jolene", "recipient": {"username": "manu"}, "subject": "test subject", "body": "hello there"}' -H "Content-Type: application/json" -X POST localhost:8080/messages
	r.POST("/messages", func(c *gin.Context) {
		var cm ComposedMessage
		if err := c.ShouldBindJSON(&cm); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		m := FromComposedMessage(cm)
		m.Send(db, c)
	})

	// curl -d '{"sender": "manu", "subject": "re:test subject", "body": "sorry, just saw this"}' -H "Content-Type: application/json" -X POST localhost:8080/messages/4/replies
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
		// set up response
		m := FromReplyMessage(replyMsg, mID)
		m.Reply(db, c)
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
		m, err := GetMessage(mID, db, c)
		// we've already updated the response with the error
		if err != nil {
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

		db.Where("user_name = ?", username).Find(&user)
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

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	return r
}
