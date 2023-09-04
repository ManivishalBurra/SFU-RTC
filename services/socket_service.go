package services

import (
	"encoding/json"
	"meshRTC/models"

	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"meshRTC/constants"
	"net/http"
	"time"
)

func StartWebSockets(house *models.MainHouse) {
	for {
		select {
		case s := <-house.Register:
			fmt.Println(s, "registering user")
			connections := house.Rooms[s.Room]
			users := house.Users[s.Room]
			if connections == nil {
				connections = make(map[*models.Connection]bool)
				house.Rooms[s.Room] = connections
			}
			if users == nil {
				users = make(map[string]*models.User)
				house.Users[s.Room] = users
			}
			house.Users[s.Room][s.User] = &s
			house.Rooms[s.Room][s.Conn] = true
		case s := <-house.Unregister:
			connections := house.Rooms[s.Room]
			users := house.Users[s.Room]
			if users != nil {
				if _, ok := users[s.User]; ok {
					delete(users, s.User)
					if len(users) == 0 {
						delete(house.Users, s.User)
					}
				}
			}
			if connections != nil {
				if _, ok := connections[s.Conn]; ok {
					delete(connections, s.Conn)
					close(s.Conn.Send)
					if len(connections) == 0 {
						delete(house.Rooms, s.Room)
					}
				}
			}
		case m := <-house.Broadcast:
			jsonData, err := json.Marshal(m)
			if err != nil {
				fmt.Println("Failed to marshal struct to JSON:", err)
				return
			}
			byteSlice := []byte(jsonData)
			connections := house.Rooms[m.Room]
			switch {
			case len(m.ToSpecificUser) > 0:
				u1 := house.Users[m.Room][m.ToSpecificUser]
				select {
				case u1.Conn.Send <- byteSlice:
				default:
					close(u1.Conn.Send)
					delete(connections, u1.Conn)
					if len(connections) == 0 {
						delete(house.Rooms, m.Room)
					}
				}
			default:
				for c := range connections {
					select {
					case c.Send <- byteSlice:
					default:
						close(c.Send)
						delete(connections, c)
						if len(connections) == 0 {
							delete(house.Rooms, m.Room)
						}
					}
				}
			}

		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  7035,
	WriteBufferSize: 7035,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ReadPump pumps messages from the websocket connection to the hub.
func ReadPump(s *models.User, house *models.MainHouse) {
	c := s.Conn
	defer func() {
		house.Unregister <- *s
		c.Ws.Close()
	}()
	c.Ws.SetReadLimit(constants.MaxMessageSize)
	c.Ws.SetReadDeadline(time.Now().Add(constants.PongWait))
	c.Ws.SetPongHandler(func(string) error { c.Ws.SetReadDeadline(time.Now().Add(constants.PongWait)); return nil })
	for {
		_, msg, err := c.Ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			log.Printf("error: %v", err)
			break
		}
		var m models.Message
		err = json.Unmarshal(msg, &m)
		if err != nil {
			fmt.Println("Failed to unmarshal JSON:", err)
			return
		}

		house.Broadcast <- m
	}
}

// Write writes a Message with the given Message type and payload.
func Write(mt int, payload []byte, c *models.Connection) error {
	c.Ws.SetWriteDeadline(time.Now().Add(constants.WriteWait))
	return c.Ws.WriteMessage(mt, payload)
}

// WritePump pumps messages from the hub to the websocket connection.
func WritePump(s *models.User) {
	c := s.Conn
	ticker := time.NewTicker(constants.PingPeriod)
	defer func() {
		ticker.Stop()
		c.Ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				Write(websocket.CloseMessage, []byte{}, c)
				return
			}
			if err := Write(websocket.TextMessage, message, c); err != nil {
				return
			}
		case <-ticker.C:
			if err := Write(websocket.PingMessage, []byte{}, c); err != nil {
				return
			}
		}
	}
}

// ServeWs handles websocket requests from the peer.
func ServeWs(w http.ResponseWriter, r *http.Request, roomId string, userId string, house *models.MainHouse) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	fmt.Println("upgrader setup correctly")
	newConnection := &models.Connection{Send: make(chan []byte, 256), Ws: ws}
	newUser := models.User{Conn: newConnection, Room: roomId, User: userId}
	house.Register <- newUser
	fmt.Println("created a user with RoomId and userId", roomId, userId)
	fmt.Println(house, "house")
	go WritePump(&newUser)
	go ReadPump(&newUser, house)
}
