package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type client struct {
	conn *websocket.Conn
	send chan []byte
}

var (
	clients   = make(map[*client]bool)
	broadcast = make(chan []byte)
	lock      sync.Mutex
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade:", err)
		return
	}
	defer conn.Close()

	c := &client{conn: conn, send: make(chan []byte)}
	lock.Lock()
	clients[c] = true
	lock.Unlock()

	go func() {
		for msg := range c.send {
			err := conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("Write Error:", err)
				break
			}
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read Error:", err)
			break
		}

		lock.Lock()
		for cli := range clients {
			if cli != c {
				cli.send <- msg
			}
		}
		lock.Unlock()
	}

	lock.Lock()
	delete(clients, c)
	close(c.send)
	lock.Unlock()
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)

	log.Println("Server started at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
