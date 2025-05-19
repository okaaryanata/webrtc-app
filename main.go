package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// limit 2 peers
var peers = make([]*websocket.Conn, 0, 2)

func main() {
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", handleWebSocket)

	addr := ":8080"
	log.Printf("Server started at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html")
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	if len(peers) >= 2 {
		log.Println("Max peers connected, rejecting new connection")
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"Max peers connected"}`))
		return
	}

	peers = append(peers, conn)
	log.Println("New peer connected:", len(peers))

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		for _, peer := range peers {
			if peer != conn {
				err = peer.WriteMessage(mt, message)
				if err != nil {
					log.Println("Write error:", err)
				}
			}
		}
	}

	for i, peer := range peers {
		if peer == conn {
			peers = append(peers[:i], peers[i+1:]...)
			break
		}
	}
	log.Println("Peer disconnected:", len(peers))
}
