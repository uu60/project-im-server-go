package server

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
		u.GetMessage(msg)
	}
}

// Online 用户的上线业务
func (u *User) Online() {
	s := u.server
	// 用户上线，将用户加入到onlineMap中
	s.MapLock.Lock()
	s.OnlineMap[u.Name] = u
	s.MapLock.Unlock()

	// 广播当前用户上线信息
	s.BroadCast(u, "已上线")
}

// Offline 用户的下线业务
func (u *User) Offline() {
	s := u.server
	// 用户上线，将用户加入到onlineMap中
	s.MapLock.Lock()
	delete(s.OnlineMap, u.Name)
	s.MapLock.Unlock()

	// 广播当前用户上线信息
	s.BroadCast(u, "已下线")
}

// SendMessage 用户发送消息的业务
func (u *User) SendMessage(msg string) {
	server := u.server
	if msg == "who" {
		// 查询当前在线用户
		server.MapLock.RLock()
		for _, user := range server.OnlineMap {
			u.GetMessage("在线: [" + user.Addr + "]" + user.Name)
		}
		server.MapLock.RUnlock()
	} else if strings.HasPrefix(msg, "rename ") {
		if !u.rename(msg, server) {
			return
		}
	} else if strings.HasPrefix(msg, "send->") {
		if !u.sendPrivateMessage(msg) {
			return
		}
	} else {
		// 普通广播消息
		server.BroadCast(u, msg)
	}
}

// sendPrivateMessage 发送私聊
func (u *User) sendPrivateMessage(msg string) bool {
	after, _ := strings.CutPrefix(msg, "send->")
	if strings.HasPrefix(after, " ") {
		u.GetMessage("失败: 请输入用户名")
		return false
	} else if len(strings.Split(after, " ")[0]) > 10 {
		u.GetMessage("失败: 用户名不能超过10个字符")
		return false
	} else {
		receiverName := strings.Split(after, " ")[0]
		user, ok := u.server.OnlineMap[receiverName]
		if ok {
			msg, _ := strings.CutPrefix(after, receiverName+" ")
			msg = GetPrefixedMessage(u, msg, true)
			user.GetMessage(msg)
			u.GetMessage(msg)
			return true
		} else {
			u.GetMessage("失败: 该用户不在线")
			return false
		}
	}
}

// rename 修改用户名
func (u *User) rename(msg string, server *Server) bool {
	after, _ := strings.CutPrefix(msg, "rename ")
	if strings.Contains(after, " ") {
		u.GetMessage("失败: 用户名不能包含空格")
		return false
	} else if len(after) > 10 {
		u.GetMessage("失败: 用户名不能超过10个字符")
		return false
	}
	// 查询是否已经存在
	server.MapLock.Lock()
	newName := after
	_, ok := server.OnlineMap[newName]
	if ok {
		u.GetMessage("失败: 用户名已存在")
		return false
	}
	delete(server.OnlineMap, u.Name)
	u.Name = newName
	server.OnlineMap[newName] = u
	server.MapLock.Unlock()
	u.GetMessage("成功: 用户名已更改为[" + newName + "]")
	return true
}

// GetMessage 给指定用户发消息
func (u *User) GetMessage(msg string) {
	u.conn.Write([]byte("\r" + msg + "\n"))
	u.conn.Write([]byte("发送: "))
}
