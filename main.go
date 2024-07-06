package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type User struct {
	ID   uint
	Name string
	Conn *websocket.Conn
}

type Channel struct {
	Subscribers map[*websocket.Conn]*User
}

var chatChannel = &Channel{
	Subscribers: make(map[*websocket.Conn]*User),
}

type Message struct {
	UserID   uint   `json:"user_id"`
	UserName string `json:"user_name"`
	Content  string `json:"content"`
}

func (c *Channel) Broadcast(message Message, excludeSender bool) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Println("error: %v", err)
		return
	}

	for conn, user := range c.Subscribers {
		if excludeSender && user.ID == message.UserID {
			continue
		}

		err := conn.WriteMessage(websocket.TextMessage, messageBytes)
		if err != nil {
			log.Println("error: %v", err)
			conn.Close()
			delete(c.Subscribers, conn)
		}
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error: %v", err)
		return
	}
	defer conn.Close()

	uID, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		log.Println("error: %v", err)
		return
	}

	user := &User{
		ID:   uint(uID),
		Name: "user1",
		Conn: conn,
	}

	chatChannel.Subscribers[conn] = user

	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			log.Println("error: %v", err)
			delete(chatChannel.Subscribers, conn)
			break
		}

		formattedMessage := Message{
			UserID:   user.ID,
			UserName: user.Name,
			Content:  string(messageBytes),
		}

		chatChannel.Broadcast(formattedMessage, false)
	}
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
