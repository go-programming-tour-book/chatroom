package logic

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var globalUID uint32 = 0

type User struct {
	UID            int           `json:"uid"`
	NickName       string        `json:"nickname"`
	EnterAt        time.Time     `json:"enter_at"`
	Addr           string        `json:"addr"`
	MessageChannel chan *Message `json:"-"`

	conn *websocket.Conn
}

// 系统用户，代表是系统主动发送的消息
var System = &User{}

func NewUser(conn *websocket.Conn, nickname, addr string) *User {
	return &User{
		UID:            int(atomic.AddUint32(&globalUID, 1)),
		NickName:       nickname,
		Addr:           addr,
		EnterAt:        time.Now(),
		MessageChannel: make(chan *Message, 8),

		conn: conn,
	}
}

func (u *User) String() string {
	return "UID:" + strconv.Itoa(u.UID) + ";nickname:" + u.NickName + ";" +
		u.EnterAt.Format("2006-01-02 15:04:05 +8000") + " 进入聊天室"
}

func (u *User) SendMessage(ctx context.Context) {
	for msg := range u.MessageChannel {
		wsjson.Write(ctx, u.conn, msg)
	}
}

// CloseMessageChannel 避免 goroutine 泄露
func (u *User) CloseMessageChannel() {
	close(u.MessageChannel)
}

func (u *User) ReceiveMessage(ctx context.Context) error {
	var (
		receiveMsg map[string]string
		err        error
	)
	for {
		err = wsjson.Read(ctx, u.conn, &receiveMsg)
		if err != nil {
			return err
		}

		// 内容发送到聊天室
		sendMsg := NewMessage(u, receiveMsg["content"])
		Broadcaster.MessageChannel() <- sendMsg
	}
}
