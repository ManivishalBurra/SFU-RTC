package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v3"
	"io"
	"meshRTC/models"
	"time"
)

var MasterPeerConnectionMappedToUser = make(map[string]*webrtc.PeerConnection)

type StreamInfo struct {
	AudioStream *webrtc.TrackRemote
	VideoStream *webrtc.TrackRemote
	Receiver    *webrtc.RTPReceiver
}

var RemotePeerInfo = []StreamInfo{}

func SetUpMasterPeerConnection() {
	time.Sleep(5 * time.Second)
	//Create a WebSocketConnection
	ws1, u1, ch, err := ConnectTheClientToSocket("webRTC", "Master")
	if err != nil {
		fmt.Println(err)
	}

	//Reads a message that we receive and Handle it based on type ex: sdp-offer, answer, ice-candidates
	go func() {
		for {
			select {
			case msg := <-ch:
				var m models.Message
				err := json.Unmarshal(msg, &m)
				if err != nil {
					fmt.Println(err)
				}
				switch m.Type {
				case "sdp-offer":
					peerConnection, _ := PrepareNewPeerConnection(ws1, u1, m)
					data, ok := m.Data.(string)
					if !ok {
						// Handle the case where msg.Data is not of type SessionDescription
						fmt.Println("msg.Data is not of type SessionDescription")
					} else {
						// Now you can use 'data' as a SessionDescription
						d := webrtc.SessionDescription{
							SDP:  data,
							Type: 1,
						}
						err := peerConnection.SetRemoteDescription(d)
						if err != nil {
							fmt.Println(err)
						}
					}
					answer, err := peerConnection.CreateAnswer(nil)
					if err != nil {
						fmt.Println(err)
					}
					err = peerConnection.SetLocalDescription(answer)
					if err != nil {
						fmt.Println(err)
					}
					MasterPeerConnectionMappedToUser[m.User] = peerConnection
					WriteMessage(ws1, models.Message{
						Data:           answer,
						Type:           "sdp-answer",
						ToSpecificUser: m.User,
						User:           u1.User,
						Room:           u1.Room,
					})
				case "ice-candidate":
					peerConnection := MasterPeerConnectionMappedToUser[m.User]
					candidateJSON := m.Data.(string)

					var candidateInit webrtc.ICECandidateInit
					if err := json.Unmarshal([]byte(candidateJSON), &candidateInit); err != nil {
						fmt.Println(err)
					}

					err := peerConnection.AddICECandidate(candidateInit)
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	}()
}

func PrepareNewPeerConnection(ws1 *websocket.Conn, u1 *models.User, msg models.Message) (*webrtc.PeerConnection, error) {

	//create a new WebRTC API
	mediaEngineParams := &webrtc.MediaEngine{}

	//Setting mediaEnginesParams Default Codecs that Pion provides
	err := mediaEngineParams.RegisterDefaultCodecs()
	if err != nil {
		fmt.Println(err)
	}
	i := &interceptor.Registry{}
	// Use the default set of Interceptors
	if err := webrtc.RegisterDefaultInterceptors(mediaEngineParams, i); err != nil {
		panic(err)
	}

	//NewAPI Creates a new API object for keeping semi-global settings to WebRTC objects
	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngineParams))

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	// Allow us to receive 1 video track
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}

	// Create an audio track from the microphone
	//mediaStream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{})
	//
	//// Add the audio tracks to the PeerConnection
	//audioTracks := mediaStream.GetTracks()
	//
	//for _, audioTrack := range audioTracks {
	//	_, err = peerConnection.AddTrack(audioTrack)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//}

	peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Println(string(msg.Data))
			dataChannel.SendText("I received")
		})
	})

	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			// Send the ICE candidate to the remote peer
			WriteMessage(ws1, models.Message{
				Data:           candidate,
				Type:           "ice-candidate",
				ToSpecificUser: msg.User,
				User:           u1.User,
				Room:           u1.Room,
			})
		} else {
			return
		}
	})

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State has changed: %s\n", connectionState.String())
	})

	// Set up the OnTrack event handler to handle incoming tracks
	peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Printf("Received a remote track: %s\n", remoteTrack.ID())
		// Handle the incoming track here (e.g., play the audio or display the video)
		localTrackHelper, err := webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, remoteTrack.ID(), remoteTrack.StreamID())
		if err != nil {
			fmt.Println(err)
		}

		_, err = peerConnection.AddTrack(localTrackHelper)
		if err != nil {
			return
		}

		rtpBuf := make([]byte, 1400)
		for {
			i, _, readErr := remoteTrack.Read(rtpBuf)
			if readErr != nil {
				panic(readErr)
			}

			// ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
			if _, err = localTrackHelper.Write(rtpBuf[:i]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
				panic(err)
			}
		}

		//RemotePeer := StreamInfo{}
		//
		//RemotePeer.Receiver = receiver
		//
		//if track.Kind() == webrtc.RTPCodecTypeAudio {
		//	RemotePeer.AudioStream = track
		//} else if track.Kind() == webrtc.RTPCodecTypeVideo {
		//	RemotePeer.VideoStream = track
		//}
		//
		//RemotePeerInfo = append(RemotePeerInfo, RemotePeer)
		//
		//fmt.Println(peerConnection.ConnectionState(), "connection state <-")

		//for _, mappedPeerWithMaster := range MasterPeerConnectionMappedToUser {
		//	trackLocal, err := webrtc.NewTrackLocalStaticRTP(track.Codec().RTPCodecCapability, track.ID(), track.StreamID())
		//	if err != nil {
		//		fmt.Println(err)
		//	}
		//	_, err = mappedPeerWithMaster.AddTrack(trackLocal)
		//	if err != nil {
		//		fmt.Println(err)
		//	}
		//}

	})

	return peerConnection, nil
}

func createAudioTrack() (*webrtc.TrackLocalStaticRTP, error) {
	// Create audio track with Opus codec
	track, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion")
	if err != nil {
		return nil, err
	}

	go func() {
		sampleRate := 48000
		pcmSize := sampleRate / 50 // 20ms of audio at 48kHz

		for {
			// Simulate audio data (silence)
			audioData := make([]byte, pcmSize*2) // 16-bit PCM
			_, err := track.Write(audioData)
			if err != nil {
				fmt.Printf("Error writing audio data to track: %v\n", err)
			}

			// Simulate a 20ms delay (adjust as needed for your audio source)
			time.Sleep(20 * time.Millisecond)
		}
	}()

	return track, nil
}
