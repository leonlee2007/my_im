package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	MapLock   sync.RWMutex
	Message   chan string
}

func (server *Server) BroadCast(user *User, sendMsg string) {
	for _, user := range server.OnlineMap {
		user.C <- sendMsg
	}
}

//处理连接
func (server *Server) Handler(conn net.Conn) {
	fmt.Printf("建立连接conn: %v\n", conn)
	defer fmt.Println("handle done...")
	//创建用户
	user := NewUser(conn, server)
	fmt.Printf("创建User: %v\n", user.Name)
	user.Online()
	go user.Process()
	for {
		select {
		case <-user.IsAlive:
		case <-time.After(time.Second * 600):
			println("timeout...")
			user.sendMsg("you are kicked...")
			// close(user.C)
			// close(user.IsAlive)
			conn.Close()
			return //runtime.Goexit()
		}
	}
}

func (server *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	defer listener.Close()
	for {
		conn, err2 := listener.Accept()
		if err2 != nil {
			fmt.Printf("err2: %v\n", err2)
			continue
		}
		go server.Handler(conn)
	}
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}
