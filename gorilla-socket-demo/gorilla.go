package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

// Client struct to hold the connection and the unique name
type Client struct {
	conn *websocket.Conn
	name string
}

var clients []Client //slice of client connections
var nextID int = 1 // Counter to assign unique names

func main() {
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Error upgrading WebSocket:", err)
			return
		}

		// Assign a unique name to each client
		clientName := "Client" + strconv.Itoa(nextID)
		nextID++

		// Store the connection and the name
		clients = append(clients, Client{conn, clientName})

		defer func() {
			conn.Close()
			// Remove client on disconnection
			for i, client := range clients {
				if client.conn == conn {
					clients = append(clients[:i], clients[i+1:]...)
					break
				}
			}
		}()

		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Printf("%s error: %v\n", clientName, err)
				break
			}

			// Add the client's name to the message
			msgWithSender := []byte(clientName + ": " + string(msg))

			fmt.Printf("%s send: %s\n", clientName, string(msg))

			// Send the modified message to all clients
			for _, client := range clients {
				if err = client.conn.WriteMessage(msgType, msgWithSender); err != nil {
					fmt.Printf("Error sending to %s: %v\n", client.name, err)
				}
			}
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	fmt.Println("Server running on 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
