package logic

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
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
	Token          string        `json:"token"`

	conn *websocket.Conn

	isNew bool
}

// 系统用户，代表是系统主动发送的消息
var System = &User{}

func NewUser(conn *websocket.Conn, token, nickname, addr string) *User {
	user := &User{
		NickName:       nickname,
		Addr:           addr,
		EnterAt:        time.Now(),
		MessageChannel: make(chan *Message, 32),
		Token:          token,

		conn: conn,
	}

	if user.Token != "" {
		uid, err := parseTokenAndValidate(token, nickname)
		if err == nil {
			user.UID = uid
		}
	}

	if user.UID == 0 {
		user.UID = int(atomic.AddUint32(&globalUID, 1))
		user.Token = genToken(user.UID, user.NickName)
		user.isNew = true
	}

	return user
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
			// 判定连接是否关闭了，正常关闭，不认为是错误
			var closeErr websocket.CloseError
			if errors.As(err, &closeErr) {
				return nil
			} else if errors.Is(err, io.EOF) {
				return nil
			}

			return err
		}

		// 内容发送到聊天室
		sendMsg := NewMessage(u, receiveMsg["content"], receiveMsg["send_time"])
		sendMsg.Content = FilterSensitive(sendMsg.Content)

		// 解析 content，看看 @ 谁了
		reg := regexp.MustCompile(`@[^\s@]{2,20}`)
		sendMsg.Ats = reg.FindAllString(sendMsg.Content, -1)

		Broadcaster.Broadcast(sendMsg)
	}
}

func genToken(uid int, nickname string) string {
	secret := viper.GetString("token-secret")
	message := fmt.Sprintf("%s%s%d", nickname, secret, uid)

	messageMAC := macSha256([]byte(message), []byte(secret))

	return fmt.Sprintf("%suid%d", base64.StdEncoding.EncodeToString(messageMAC), uid)
}

func parseTokenAndValidate(token, nickname string) (int, error) {
	pos := strings.LastIndex(token, "uid")
	messageMAC, err := base64.StdEncoding.DecodeString(token[:pos])
	if err != nil {
		return 0, err
	}
	uid := cast.ToInt(token[pos+3:])

	secret := viper.GetString("token-secret")
	message := fmt.Sprintf("%s%s%d", nickname, secret, uid)

	ok := validateMAC([]byte(message), messageMAC, []byte(secret))
	if ok {
		return uid, nil
	}

	return 0, errors.New("token is illegal")
}

func macSha256(message, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(message)
	return mac.Sum(nil)
}

func validateMAC(message, messageMAC, secret []byte) bool {
	mac := hmac.New(sha256.New, secret)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}
