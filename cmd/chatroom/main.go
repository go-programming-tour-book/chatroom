package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/polaris1119/chatroom/server"
)

var (
	addr   = ":2022"
	banner = `
    ____              _____
   |    |    |   /\     |
   |    |____|  /  \    | 
   |    |    | /----\   |
   |____|    |/      \  |

Go语言之旅 —— 一起用Go做项目：ChatRoom，start on：%s
`
)

func main() {
	fmt.Printf(banner+"\n", addr)

	server.RegisterHandle()

	log.Fatal(http.ListenAndServe(addr, nil))
}
