package main

import (
	"net"
	"strings"
)

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
		u.ReceiveMessage(msg)
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

// SendToServerMessage 用户发送消息的业务
func (u *User) SendToServerMessage(msg string) {
	server := u.server
	if msg == "who" {
		// 查询当前在线用户
		server.mapLock.RLock()
		for _, user := range server.OnlineMap {
			u.ReceiveMessage("在线: [" + user.Addr + "]" + user.Name)
		}
		server.mapLock.RUnlock()
	} else if strings.HasPrefix(msg, "rename ") {
		if !u.rename(msg, server) {
			return
		}
	} else {
		// 普通广播消息
		server.BroadCast(u, msg)
	}
}

// rename 修改用户名
func (u *User) rename(msg string, server *Server) bool {
	after, _ := strings.CutPrefix(msg, "rename ")
	if strings.Contains(after, " ") {
		u.ReceiveMessage("失败: 用户名不能包含空格")
		return false
	} else if len(after) > 10 {
		u.ReceiveMessage("失败: 用户名不能超过10个字符")
		return false
	}
	// 查询是否已经存在
	server.mapLock.Lock()
	newName := after
	_, ok := server.OnlineMap[newName]
	if ok {
		u.ReceiveMessage("失败: 用户名已存在")
		return false
	}
	delete(server.OnlineMap, u.Name)
	u.Name = newName
	server.OnlineMap[newName] = u
	server.mapLock.Unlock()
	u.ReceiveMessage("成功: 用户名已更改为[" + newName + "]")
	return true
}

// ReceiveMessage 给指定客户发消息
func (u *User) ReceiveMessage(msg string) {
	u.conn.Write([]byte("\r" + msg + "\n"))
	u.conn.Write([]byte("发送: "))
}
