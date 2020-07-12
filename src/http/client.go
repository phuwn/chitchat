package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
)

var hub *broadcast

type broadcast struct {
	thread []string
	mux    sync.Mutex
}

func (b *broadcast) sendMessage(msg string) {
	b.mux.Lock()
	b.thread = append(b.thread, msg)
	b.mux.Unlock()
}

func (b *broadcast) getMessages(i int) (int, []string) {
	if i > len(b.thread) {
		return len(b.thread), []string{}
	}
	return len(b.thread), b.thread[i:]
}

func init() {
	hub = &broadcast{
		thread: make([]string, 0),
	}
}

// GetMessages responses a set of new messages from the provided time stamp of a client
func GetMessages(w http.ResponseWriter, r *http.Request) {
	if hub == nil {
		http.Error(w, "unknown error", http.StatusInternalServerError)
		return
	}

	iParams := r.URL.Query()["index"]
	if len(iParams) == 0 {
		http.Error(w, "missing index", http.StatusBadRequest)
		return
	}

	i, err := strconv.Atoi(iParams[0])
	if err != nil {
		http.Error(w, "timestamp has to be a number", http.StatusBadRequest)
		return
	}

	ni, msg := hub.getMessages(i)
	b, err := json.Marshal(map[string]interface{}{"index": ni, "messages": msg})
	if err != nil {
		http.Error(w, "unknown error", http.StatusInternalServerError)
		return
	}

	w.Write(b)
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

// Start begins a client session
func Start(w http.ResponseWriter, r *http.Request) {
	if hub == nil {
		http.Error(w, "unknown error", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(strconv.Itoa(len(hub.thread))))
}
