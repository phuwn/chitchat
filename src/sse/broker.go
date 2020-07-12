package sse

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var broker *Broker

// A Broker holds open client connections,
// listens for incoming events on its Notifier channel
// and broadcast event data to all registered connections
type Broker struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

// NewBroker - create and run a new broker
func NewBroker() {
	broker = &Broker{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
	broker.listen()
}

func (b *Broker) listen() {
	for {
		select {
		case client := <-b.register:
			b.clients[client] = true
		case client := <-b.unregister:
			if _, ok := b.clients[client]; ok {
				close(client.send)
				delete(b.clients, client)
			}
			log.Printf("Removed client. %d registered clients", len(b.clients))

		case event := <-b.broadcast:
			for client := range b.clients {
				select {
				case client.send <- event:
				default:
					close(client.send)
					delete(b.clients, client)
				}

			}
		}
	}
}

// SendMessage init new message to broadcast's thread
func SendMessage(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unsupport body type", http.StatusBadRequest)
		return
	}

	if broker == nil {
		http.Error(w, "unknown error", http.StatusInternalServerError)
		return
	}

	broker.broadcast <- b
}

// Serve handles server sent event requests from the peer.
func Serve(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the Broker's connections registry
	client := &Client{
		send: make(chan []byte),
	}

	// Signal the broker that we have a new connection
	broker.register <- client

	// Remove this client from the map of connected clients
	// when this handler exits.
	defer func() {
		broker.unregister <- client
	}()

	// Listen to connection close and un-register messageChan
	// notify := w.(http.CloseNotifier).CloseNotify()
	notify := r.Context().Done()

	go func() {
		<-notify
		broker.unregister <- client
	}()

	for {
		fmt.Fprintf(w, "data: %s\n\n", <-client.send)

		// Flush the data immediatly instead of buffering it for later.
		flusher.Flush()
	}
}
