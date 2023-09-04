package constants

import "time"

const (
	// WriteWait Time allowed to write a Message to the peer.
	WriteWait = 10 * time.Second

	// PongWait Time allowed to read the next pong Message from the peer.
	PongWait = 60 * time.Second

	// PingPeriod Send pings to peer with this period. Must be less than pongWait.
	PingPeriod = (PongWait * 9) / 10

	// MaxMessageSize Maximum Message size allowed from peer.
	MaxMessageSize = 7035
)
