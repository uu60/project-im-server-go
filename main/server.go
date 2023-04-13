package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int
	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	BroadcastChan chan string
}

func NewServer(ip string, port int) *Server {
	// 创建一个基本的对象
	server := Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		//mapLock:       sync.RWMutex{},
		BroadcastChan: make(chan string),
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
	go s.ListenBroadcastMessage()

	// loop accept 处理新的连接请求
	for {
		// accept
		accept, err := listen.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		// do handler 业务回调
		go s.HandleNewConnection(accept)
	}

}

// HandleNewConnection 处理当前连接请求
func (s *Server) HandleNewConnection(conn net.Conn) {
	//fmt.Println("连接建立成功")
	user := NewUser(conn, s)
	conn.Write([]byte("欢迎来到聊天室!\n[who]命令查看在线用户\n[rename 你的用户名]来改名\n[send->接收人名字]来进行私聊\n\n"))
	user.Online()

	aliveChan := make(chan bool)
	// 接收客户端发送的消息
	go s.loopWaitForMessage(conn, user, aliveChan)

	// 实现超时踢出
	go s.loopHandleTimeout(conn, aliveChan, user)
}

func (s *Server) loopHandleTimeout(conn net.Conn, aliveChan chan bool, user *User) {
	for {
		// 监听多个管道
		select {
		case <-aliveChan:
			// 自动执行下面的来重置定时
		case <-time.After(time.Second * 60):
			// 已经超时,踢出当前user
			user.GetMessage("系统: 超时踢出")
			close(user.C)
			// 关闭连接
			conn.Close()
			return
		}
	}
}

func (s *Server) loopWaitForMessage(conn net.Conn, user *User, aliveChan chan bool) {
	buf := make([]byte, 4096)
	for {
		// 阻塞等待用户发送消息
		read, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("Conn read err:", err)
		}

		// 发送空白字符长度为1
		if read == 0 {
			user.Offline()
			return
		}

		// 提取用户的消息 去除\n
		msg := string(buf[:read-1])
		// 将得到的消息进行广播
		user.SendMessage(msg)
		aliveChan <- true
	}
}

func (s *Server) BroadCast(sender *User, msg string) {
	sendMsg := getPrefixedMessage(sender, msg, false)
	s.BroadcastChan <- sendMsg
}

func (s *Server) ListenBroadcastMessage() {
	for {
		msg := <-s.BroadcastChan

		// 将msg发送给全部在线的user
		s.mapLock.RLock()
		for _, cli := range s.OnlineMap {
			cli.C <- msg
		}
		s.mapLock.RUnlock()
	}
}
