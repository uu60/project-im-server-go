package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int
	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

func NewServer(ip string, port int) *Server {
	// 创建一个基本的对象
	server := Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		mapLock:   sync.RWMutex{},
		Message:   make(chan string),
	}
	return &server
}

// 启动服务器的接口
func (this *Server) Start() {
	// socket listen
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
	}
	// close listener socket finally
	defer listen.Close()

	// 启动监听Message
	go this.ListenMessage()

	// loop accept
	for {
		// accept
		accept, err := listen.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		// do handler 业务回调
		go this.Handle(accept)
	}

}

func (this *Server) Handle(conn net.Conn) {
	// 处理当前连接请求
	//fmt.Println("连接建立成功")
	user := NewUser(conn)

	// 用户上线，将用户加入到onlineMap中
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	this.BroadCast(user, "已上线")
}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ": " + msg
	this.Message <- sendMsg
}

func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message

		// 将msg发送给全部在线的user
		this.mapLock.RLock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.RUnlock()
	}
}
