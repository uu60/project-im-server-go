package main

import "net"

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// NewUser 创建用户API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	go user.ListenMessage()

	return &user
}

// ListenMessage 监听当前User channel的方法，一旦有消息，就直接发送给对应客户端
func (u *User) ListenMessage() {
	// loop listen
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}

// Online 用户的上线业务
func (u *User) Online() {
	s := u.server
	// 用户上线，将用户加入到onlineMap中
	s.mapLock.Lock()
	s.OnlineMap[u.Name] = u
	s.mapLock.Unlock()

	// 广播当前用户上线信息
	s.BroadCast(u, "已上线")
}

// Offline 用户的下线业务
func (u *User) Offline() {
	s := u.server
	// 用户上线，将用户加入到onlineMap中
	s.mapLock.Lock()
	delete(s.OnlineMap, u.Name)
	s.mapLock.Unlock()

	// 广播当前用户上线信息
	s.BroadCast(u, "已下线")
}

// DoMessage 用户处理消息的业务
func (u *User) DoMessage(msg string) {
	server := u.server
	if msg == "who" {
		// 查询当前在线用户
		server.mapLock.Lock()
		for _, user := range server.OnlineMap {
			u.SendMsg("在线：[" + user.Addr + "]" + user.Name + "\n")
		}
	} else {
		server.BroadCast(u, msg)
	}
}

// SendMsg 给指定客户发消息
func (u *User) SendMsg(msg string) {
	u.conn.Write([]byte(msg))
}
