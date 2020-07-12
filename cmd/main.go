package main

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	usualHTTPReq "github.com/phuwn/chitchat/src/http"
	"github.com/phuwn/chitchat/src/lp"
	"github.com/phuwn/chitchat/src/sse"
	"github.com/phuwn/chitchat/src/ws"
	"github.com/phuwn/tools/log"
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if page := r.URL.Query()["page"]; len(page) != 0 {
		switch page[0] {
		case "ws":
			http.ServeFile(w, r, "page/ws.html")
			return
		case "lp":
			http.ServeFile(w, r, "page/lp.html")
			return
		case "sse":
			http.ServeFile(w, r, "page/sse.html")
			return
		}
	}
	http.ServeFile(w, r, "page/http.html")
}

func main() {
	addr := ":8080"
	log.Status("listening on port%s", addr)
	go ws.NewHub()
	go sse.NewBroker()

	r := mux.NewRouter()
	r.HandleFunc("/", serveHome)
	r.HandleFunc("/ws", ws.Serve)
	r.HandleFunc("/http", usualHTTPReq.Start).Methods("GET")
	r.HandleFunc("/http/msg", usualHTTPReq.GetMessages).Methods("GET")
	r.HandleFunc("/http/msg", usualHTTPReq.SendMessage).Methods("POST")
	r.HandleFunc("/lp", lp.PollResponse).Methods("GET")
	r.HandleFunc("/lp/msg", lp.SendMessage).Methods("POST")
	r.HandleFunc("/sse", sse.Serve)
	r.HandleFunc("/sse/msg", sse.SendMessage).Methods("POST")

	err := http.ListenAndServe(addr, handlers.CORS()(r))
	if err != nil {
		log.Status("server got terminated, err: %s", err.Error())
	}
}
