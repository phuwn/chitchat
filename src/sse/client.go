package sse

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	// Buffered channel of outbound messages.
	send chan []byte
}
