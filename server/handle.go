package server

import (
	"net/http"

	"github.com/polaris1119/chatroom/logic"
)

func RegisterHandle() {
	inferRootDir()

	// 广播消息处理
	go logic.Broadcaster.Broadcast()

	http.HandleFunc("/", homeHandleFunc)

	http.HandleFunc("/ws", WebSocketHandleFunc)
}
