package services

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"meshRTC/models"
)

func ConnectTheClientToSocket(roomId string, userId string) (*websocket.Conn, *models.User, chan []byte, error) {
	u := "ws://localhost:6303/ws/" + roomId + "/" + userId
	user := models.User{
		Room: roomId,
		User: userId,
	}
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("WebSocket connection error:", err)
		return nil, nil, nil, err
	}
	ch := make(chan []byte)
	// Receive messages from the server
	go ReadMessage(c, ch)
	return c, &user, ch, nil
}

func ReadMessage(c *websocket.Conn, ch chan []byte) {
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Fatal("Error reading message:", err)
			return
		}
		ch <- message
	}
}

func WriteMessage(c *websocket.Conn, msg models.Message) {

	jsonData, _ := json.Marshal(msg)
	err := c.WriteMessage(websocket.TextMessage, []byte(jsonData))
	if err != nil {
		fmt.Println("Error sending message:", err)
	}
}
