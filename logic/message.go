package logic

import "time"

const (
	MsgTypeNormal   = iota // 普通用户消息
	MsgTypeSystem          // 系统消息
	MsgTypeError           // 错误消息
	MsgTypeUserList        // 发送当前用户列表
)

// 给用户发送的消息
type Message struct {
	User    *User     `json:"user"`
	Type    int       `json:"type"`
	Content string    `json:"content"`
	MsgTime time.Time `json:"msg_time"`

	Users map[string]*User `json:"users"`
}

func NewMessage(user *User, content string) *Message {
	return &Message{
		User:    user,
		Type:    MsgTypeNormal,
		Content: content,
		MsgTime: time.Now(),
	}
}

func NewWelcomeMessage(nickname string) *Message {
	return &Message{
		User:    System,
		Type:    MsgTypeSystem,
		Content: nickname + " 您好，欢迎加入聊天室！",
		MsgTime: time.Now(),
	}
}

func NewNoticeMessage(content string) *Message {
	return &Message{
		User:    System,
		Type:    MsgTypeSystem,
		Content: content,
		MsgTime: time.Now(),
	}
}

func NewErrorMessage(content string) *Message {
	return &Message{
		User:    System,
		Type:    MsgTypeError,
		Content: content,
		MsgTime: time.Now(),
	}
}

func NewUserListMessage(users map[string]*User) *Message {
	return &Message{
		User:    System,
		Type:    MsgTypeUserList,
		MsgTime: time.Now(),
		Users:   users,
	}
}
