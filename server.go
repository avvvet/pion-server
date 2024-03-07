package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{}
	clients  = make(map[*websocket.Conn]bool)
	mutex    = sync.Mutex{}
)

func autoBroadcast() {
	if len(clients) > 0 {
		broadcast([]byte("hellow peer " + time.Now().String()))
		log.Println("auto broadcast sent ..")
	}
}

func main() {
	// HTTP handler for WebSocket connection
	http.HandleFunc("/ws", handleWebSocket)

	// Start HTTP server
	log.Println("Signaling server started at :30001")
	log.Fatal(http.ListenAndServe(":30001", nil))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	//defer conn.Close()

	// Add client to the list of connected clients
	mutex.Lock()
	clients[conn] = true
	mutex.Unlock()

	// Read message from client
	// _, msg, err := conn.ReadMessage()
	// if err != nil {
	// 	log.Println("Error reading message:", err)
	// 	return
	// }

	// // Log received message
	// log.Printf("Received message from client: %s\n", msg)

	// // Determine message type
	// var messageType struct {
	// 	Type string `json:"type"`
	// }
	// err = json.Unmarshal(msg, &messageType)
	// if err != nil {
	// 	log.Println("Error parsing message type:", err)
	// 	return
	// }

	// switch messageType.Type {
	// case "offer":
	// 	// Handle offer message
	// 	handleOfferMessage(conn, msg)
	// default:
	// 	// Broadcast non-offer message to all connected clients
	// 	broadcast(msg)
	// }

	ticker := time.NewTicker(5000 * time.Millisecond)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("broadcast at ", t)
				autoBroadcast()
			}
		}
	}()
}

func handleOfferMessage(sender *websocket.Conn, msg []byte) {
	// Parse offer message
	// Assuming offer message is in JSON format
	var offer struct {
		Type string `json:"type"`
		// Add other fields as needed for SDP and ICE candidates
	}
	err := json.Unmarshal(msg, &offer)
	if err != nil {
		log.Println("Error parsing offer message:", err)
		return
	}

	// Forward offer message to all other connected clients
	mutex.Lock()
	defer mutex.Unlock()
	for conn := range clients {
		if conn != sender {
			err := conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("Error forwarding offer message:", err)
			}
		}
	}
}

func broadcast(msg []byte) {
	mutex.Lock()
	defer mutex.Unlock()

	for conn := range clients {
		err := conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("Error writing message:", err)
			return
		}
	}
}
