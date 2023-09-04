package tests

import (
	"encoding/json"
	"fmt"
	"meshRTC/models"
	"meshRTC/services"
	"testing"
	"time"
)

func TestTwoClientConnection(t *testing.T) {

	//Creating Client Connection to Server
	c1, u1, c1ReadMessageChan, err := services.ConnectTheClientToSocket("123", "456")
	if err != nil {
		t.Error(err)
	}
	c2, u2, c2ReadMessageChan, err := services.ConnectTheClientToSocket("123", "789")
	if err != nil {
		t.Error(err)
	}

	go func() {
		var x models.Message
		for {
			select {
			case readMessagec1 := <-c1ReadMessageChan:
				json.Unmarshal(readMessagec1, &x)
				if x.Room != "123" || (len(x.ToSpecificUser) > 0 && x.ToSpecificUser != u1.User) {
					t.Error("Message came from another Room or specificUser issue", x, u1)
				}
				fmt.Println(x.User, x.ToSpecificUser, u1.User)
			case readMessagec2 := <-c2ReadMessageChan:
				json.Unmarshal(readMessagec2, &x)
				if x.Room != "123" || (len(x.ToSpecificUser) > 0 && x.ToSpecificUser != u2.User) {
					t.Error("Message came from another Room or specificUser issue", x, u2)
				}
				fmt.Println(x.User, x.ToSpecificUser, u2.User)
			}
		}
	}()

	// Send a message to the server
	services.WriteMessage(c1, models.Message{
		Data:           "Hello this is 456 from room 123",
		Room:           u1.Room,
		User:           u1.User,
		ToSpecificUser: "",
		ProfilePic:     "",
		Type:           "",
	})
	services.WriteMessage(c2, models.Message{
		Data:           "Hello this is 789 from room 123",
		Room:           u2.Room,
		User:           u2.User,
		ToSpecificUser: "",
		ProfilePic:     "",
		Type:           "",
	})

	//Send a Message to Specific User only
	services.WriteMessage(c1, models.Message{
		Data:           "Hello this is " + u1.User + " from room " + u1.Room,
		Room:           u1.Room,
		User:           u1.User,
		ToSpecificUser: u2.User,
		ProfilePic:     "",
		Type:           "",
	})
	services.WriteMessage(c2, models.Message{
		Data:           "Hello this is " + u2.User + " from room " + u2.Room,
		Room:           u2.Room,
		User:           u2.User,
		ToSpecificUser: u1.User,
		ProfilePic:     "",
		Type:           "",
	})

	time.Sleep(5 * time.Second)
	defer func() {
		err = c1.Close()
		err = c2.Close()
		if err != nil {
			t.Error(err)
		}
	}()
}
