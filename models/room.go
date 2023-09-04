package models

import (
	"github.com/gorilla/websocket"
)

type User struct {
	Conn *Connection
	Room string
	User string
}

// Connection is a middleman between the websocket connection and the hub.
type Connection struct {
	// The websocket connection.
	Ws *websocket.Conn

	// Buffered channel of outbound messages.
	Send chan []byte

	// User For this Connection
	User string
}

type MainHouse struct {
	// Registered connections.
	Rooms map[string]map[*Connection]bool

	//Registered users
	Users map[string]map[string]*User

	// Inbound messages from the connections.
	Broadcast chan Message

	// Register requests from the connections.
	Register chan User

	// Unregister requests from connections.
	Unregister chan User
}

type Clients struct {
	User string `json:"user"`
}

type House struct {
	Rooms []Room
}

type Room struct {
	RoomId string
	User   []Users
}

type Users struct {
	UserId string
	RoomId string
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub
}

type Hub struct {
	Users      map[*Users]bool
	Broadcast  chan []byte
	Register   chan *Users
	UnRegister chan *Users
}

type MessageData struct {
	Data   []byte
	RoomId string
	UserId string
}

type Message struct {
	Data           interface{} `json:"data"`
	Room           string      `json:"room"`
	User           string      `json:"user"`
	ToSpecificUser string      `json:"toSpecificUser"`
	ProfilePic     string      `json:"profilepic"`
	Type           string      `json:"type"`
}
