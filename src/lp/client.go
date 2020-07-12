package lp

import (
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

var hub *Hub

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	index  int
	client map[int]chan string
	mux    sync.Mutex
}

func init() {
	hub = &Hub{
		client: make(map[int]chan string),
	}
}

func (h *Hub) newClient() int {
	h.mux.Lock()
	index := h.index
	h.client[index] = make(chan string)
	h.index++
	h.mux.Unlock()
	return index
}

func (h *Hub) sendMessage(msg string) {
	for _, v := range h.client {
		v <- msg
	}
}

func (h *Hub) endSession(index int) {
	h.mux.Lock()
	close(h.client[index])
	delete(h.client, index)
	h.mux.Unlock()
}

// SendMessage init new message to broadcast's thread
func SendMessage(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unsupport body type", http.StatusBadRequest)
		return
	}

	if hub == nil {
		http.Error(w, "unknown error", http.StatusInternalServerError)
		return
	}

	hub.sendMessage(string(b))
}

// PollResponse handle longpolling request
func PollResponse(w http.ResponseWriter, r *http.Request) {
	timeout := make(chan bool)
	go func() {
		time.Sleep(30e9)
		timeout <- true
	}()

	id := hub.newClient()
	defer func() {
		close(timeout)
		hub.endSession(id)
	}()

	for {
		select {
		case msg := <-hub.client[id]:
			io.WriteString(w, msg)
			return
		case <-timeout:
			return
		}
	}
}
