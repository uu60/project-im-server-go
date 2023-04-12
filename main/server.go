package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

func NewServer(ip string, port int) *Server {
	// 创建一个基本的对象
	server := Server{Ip: ip, Port: port}
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
	fmt.Println("连接建立成功")
}
