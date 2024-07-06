
### Websocket
1. params: user_id & channel_id
   
### Improvement

```go
// user modal
// user.is_online

// On user reconnect logic
func onUserReconnect(user *User) {
	messages := RetrieveMessagesForUser(user.ID)
	for _, message := range messages {
	    user.Conn.WriteMessage(websocket.TextMessage, []byte(message.Content))
	}
}
```
