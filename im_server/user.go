package main

import (
	"fmt"
	"io"
	"net"
	"strings"
)

type User struct {
	Name    string
	Addr    string
	C       chan string
	Conn    net.Conn
	Server  *Server
	IsAlive chan bool
}

func (user *User) ListenMessage() {
	for {
		msg := <-user.C
		user.Conn.Write([]byte(msg + "\n"))
	}
}

func (user *User) sendMsg(msg string) {
	user.Conn.Write([]byte(msg + "\n"))
}

func (user *User) Online() {
	//将用户加入到OnlineMap
	server := user.Server
	server.OnlineMap[user.Name] = user
	server.BroadCast(user, fmt.Sprintf("%s上线了...", user.Name))
}

func (user *User) Offline() {
	//将用户加入到OnlineMap
	server := user.Server
	delete(server.OnlineMap, user.Name)
	server.BroadCast(user, fmt.Sprintf("%s离线了...", user.Name))
}

func (user *User) Process() {
	buf := make([]byte, 4096)
	user.loop(buf)
}

func (user *User) loop(buf []byte) {
	conn := user.Conn
	n, err := conn.Read(buf)
	server := user.Server
	if n == 0 {
		fmt.Printf("err: %v\n", err)
		fmt.Printf("name:%v 离线了...\n", user.Name)
		server.MapLock.Lock()
		user.Offline()
		server.MapLock.Unlock()
		return
	}
	if err != nil && err != io.EOF {
		server.MapLock.Lock()
		fmt.Printf("err: %v\n", err)
		server.MapLock.Unlock()
		return
	}
	msg := string(buf[:n-1])
	fmt.Printf("name: %v, msg: %v\n", user.Name, msg)
	server.MapLock.Lock()
	user.HandleMsg(msg)
	server.MapLock.Unlock()
	user.IsAlive <- true
	user.loop(buf)
}

//消息处理
//rename?新名称 修改名称
//to?在线用户名称?聊天内容 私聊
func (user *User) HandleMsg(msg string) {
	server := user.Server
	if msg == "who" {
		user.HandleWhoMsg(msg)
	} else if len(msg) > 7 && msg[:7] == "rename?" {
		user.HandleRenameMsg(msg)
	} else if len(msg) > 4 && msg[:3] == "to?" {
		user.HandlePrivateMsg(msg)
		//广播
	} else {
		sendMsg := fmt.Sprintf("[%s][广播]:%s", user.Name, msg)
		server.BroadCast(user, sendMsg)
	}
}

func (user *User) HandleWhoMsg(msg string) {
	server := user.Server
	for _, u := range server.OnlineMap {
		sendMsg := fmt.Sprintf("[%s]%s在线", u.Addr, u.Name)
		user.sendMsg(sendMsg)
	}
}

func (user *User) HandleRenameMsg(msg string) {
	server := user.Server
	newName := strings.Split(msg, "?")[1]
	if _, ok := server.OnlineMap[newName]; ok {
		user.sendMsg(fmt.Sprintf("用户名:%s已被使用", newName))
	} else {
		server.OnlineMap[newName] = user
		delete(server.OnlineMap, user.Name)
		oldName := user.Name
		user.Name = newName
		user.sendMsg(fmt.Sprintf("Old:%s->New:%s修改名字成功", oldName, newName))
	}
}

func (user *User) HandlePrivateMsg(msg string) {
	server := user.Server
	remoteName := strings.Split(msg, "?")[1]
	if remoteUser, ok := server.OnlineMap[remoteName]; !ok {
		user.sendMsg(fmt.Sprintf("用户名:%s不存在!!!", remoteName))
	} else {
		content := strings.Split(msg, "?")[2]
		remoteUser.sendMsg(fmt.Sprintf("[%s][私聊]:%s", user.Name, content))
	}
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:    userAddr,
		Addr:    userAddr,
		C:       make(chan string),
		Conn:    conn,
		Server:  server,
		IsAlive: make(chan bool),
	}
	go user.ListenMessage()
	return user
}
