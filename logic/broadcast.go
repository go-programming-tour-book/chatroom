package logic

import (
	"log"
)

const (
	MessageQueueLen = 8
)

// broadcaster 广播器
type broadcaster struct {
	users map[string]*User

	// 所有 channel 统一管理，可以避免外部乱用

	enteringChannel chan *User
	leavingChannel  chan *User
	messageChannel  chan *Message

	// 判断该昵称用户是否可进入聊天室（重复与否）：true 能，false 不能
	checkUserChannel      chan string
	checkUserCanInChannel chan bool
}

var Broadcaster = &broadcaster{
	users: make(map[string]*User),

	enteringChannel: make(chan *User),
	leavingChannel:  make(chan *User),
	messageChannel:  make(chan *Message, MessageQueueLen),

	checkUserChannel:      make(chan string),
	checkUserCanInChannel: make(chan bool),
}

func (b *broadcaster) Broadcast() {
	for {
		select {
		case user := <-b.enteringChannel:
			// 新用户进入
			b.users[user.NickName] = user

			b.sendUserList()
		case user := <-b.leavingChannel:
			// 用户离开
			delete(b.users, user.NickName)
			// 避免 goroutine 泄露
			user.CloseMessageChannel()

			b.sendUserList()
		case msg := <-b.messageChannel:
			// 给所有在线用户发送消息
			for _, user := range b.users {
				if user.UID == msg.User.UID {
					continue
				}
				user.MessageChannel <- msg
			}
		case nickname := <-b.checkUserChannel:
			if _, ok := b.users[nickname]; ok {
				b.checkUserCanInChannel <- false
			} else {
				b.checkUserCanInChannel <- true
			}
		}
	}
}

func (b *broadcaster) EnteringChannel() chan<- *User {
	return b.enteringChannel
}

func (b *broadcaster) LeavingChannel() chan<- *User {
	return b.leavingChannel
}

func (b *broadcaster) MessageChannel() chan<- *Message {
	return b.messageChannel
}

func (b *broadcaster) CheckUserChannel() chan<- string {
	return b.checkUserChannel
}

func (b *broadcaster) CheckUserCanInChannel() <-chan bool {
	return b.checkUserCanInChannel
}

func (b *broadcaster) sendUserList() {
	// 避免死锁，存在用户看到的列表没及时更新的可能性
	if len(b.messageChannel) < MessageQueueLen {
		b.messageChannel <- NewUserListMessage(b.users)
	} else {
		log.Println("消息并发量过大，导致 MessageChannel 拥堵。。。")
	}
}
