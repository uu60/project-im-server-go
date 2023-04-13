package main

import (
	"fmt"
	"io"
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
func (s *Server) Start() {
	// socket listen
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
	}
	// close listener socket finally
	defer listen.Close()

	// 启动监听Message
	go s.ListenMessage()

	// loop accept
	for {
		// accept
		accept, err := listen.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		// do handler 业务回调
		go s.Handle(accept)
	}

}

func (s *Server) Handle(conn net.Conn) {
	// 处理当前连接请求
	//fmt.Println("连接建立成功")
	user := NewUser(conn)

	// 用户上线，将用户加入到onlineMap中
	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock()

	// 广播当前用户上线信息
	s.BroadCast(user, "已上线")

	// 接收客户端发送的消息
	buf := make([]byte, 4096)
	for {
		read, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("Conn read err:", err)
		}

		if read == 0 {
			s.BroadCast(user, "下线")
			return
		}

		// 提取用户的消息 去除\n
		msg := string(buf[:read-1])
		// 将得到的消息进行广播
		s.BroadCast(user, msg)
	}
}

func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ": " + msg
	s.Message <- sendMsg
}

func (s *Server) ListenMessage() {
	for {
		msg := <-s.Message

		// 将msg发送给全部在线的user
		s.mapLock.RLock()
		for _, cli := range s.OnlineMap {
			cli.C <- msg
		}
		s.mapLock.RUnlock()
	}
}
