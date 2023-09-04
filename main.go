package main

import (
	"meshRTC/api"
	"meshRTC/models"
	"meshRTC/services"
)

func main() {
	//Create a House Struct
	var house = models.MainHouse{
		Broadcast:  make(chan models.Message),
		Register:   make(chan models.User),
		Unregister: make(chan models.User),
		Users:      make(map[string]map[string]*models.User),
		Rooms:      make(map[string]map[*models.Connection]bool),
	}

	//StartWebsocket connect read and write
	go services.StartWebSockets(&house)

	//Starting WebRTC Peer Connection
	go services.SetUpMasterPeerConnection()

	// Starting gin server on port 6303
	api.StartServer(&house)

}
