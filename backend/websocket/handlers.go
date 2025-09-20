package websocket

func handleMessage(client *Client, msg *Message) Message {
	return Message{
		Type:    "response",
		Action:  msg.Action,
		Success: false,
		Error:   "WebSocket not used for Twitter operations",
	}
}
