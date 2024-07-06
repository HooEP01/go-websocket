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
	ID       uint
	Name     string
	Conn     *websocket.Conn
	IsOnline bool
}

type Channel struct {
	ID          string
	Subscribers map[*websocket.Conn]*User
}

type ChannelManager struct {
	Channels map[string]*Channel
}

func NewChannelManager() *ChannelManager {
	return &ChannelManager{
		Channels: make(map[string]*Channel),
	}
}

func (manager *ChannelManager) CreateChannel(channelID string) *Channel {
	channel := &Channel{
		ID:          channelID,
		Subscribers: make(map[*websocket.Conn]*User),
	}
	manager.Channels[channelID] = channel
	return channel
}

func (manager *ChannelManager) GetChannel(channelID string) (*Channel, bool) {
	channel, exist := manager.Channels[channelID]
	return channel, exist
}

var channelManager = NewChannelManager()

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

	channelID := r.URL.Query().Get("channel_id")

	user := &User{
		ID:   uint(uID),
		Name: "user" + userID,
		Conn: conn,
	}

	channel, exist := channelManager.GetChannel(channelID)
	if !exist {
		channel = channelManager.CreateChannel(channelID)
	}

	channel.Subscribers[conn] = user

	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			log.Println("error: %v", err)
			delete(channel.Subscribers, conn)
			break
		}

		formattedMessage := Message{
			UserID:   user.ID,
			UserName: user.Name,
			Content:  string(messageBytes),
		}

		channel.Broadcast(formattedMessage, false)
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
