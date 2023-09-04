package tests

import (
	"encoding/json"
	"fmt"
	"meshRTC/models"
	"meshRTC/services"
	"testing"
	"time"
)

func TestClientWebRtcConnection(t *testing.T) {

	//Creating Client Connection to Server
	clientWSConn, userClient, c1Ch, err := services.ConnectTheClientToSocket("webRTC", "Gowtham")
	if err != nil {
		t.Error(err)
	}

	//Create Peer Connection For the Client
	peerConnection, err := services.PrepareNewPeerConnection(clientWSConn, userClient, models.Message{
		User: "Master",
	})
	if err != nil {
		t.Error(err)
	}

	//Create an Offer and send it to Master
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		t.Error(err)
	}
	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		t.Error(err)
	}
	services.WriteMessage(clientWSConn, models.Message{
		Data:           offer.SDP,
		User:           userClient.User,
		ToSpecificUser: "Master",
		Type:           "sdp-offer",
		Room:           userClient.Room,
	})
	go func() {
		for {
			select {
			case msg := <-c1Ch:
				var m models.Message
				err := json.Unmarshal(msg, &m)
				if err != nil {
					fmt.Println(err)
				}
				switch m.Type {
				case "ice-candidate":

				case "answer":

				}
			}
		}
	}()

	time.Sleep(5 * time.Minute)
	defer func() {
		err = clientWSConn.Close()
		if err != nil {
			t.Error(err)
		}
	}()
}
