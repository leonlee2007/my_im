package main

import (
	"fmt"
	"net"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	Conn       net.Conn
	Flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		Flag:       88,
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return nil
	}
	client.Conn = conn
	return client
}

func (client *Client) handleReply() {
	// io.Copy(os.Stdout, client.Conn)
	for {
		buff := make([]byte, 4096)
		_, err := client.Conn.Read(buff)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return
		}
		fmt.Printf(">:%s", string(buff))
	}
}

func (client *Client) privateChat() {
	remoteName := ""
	for {
		client.sendMsgToServer("who\n")
		fmt.Println(">>>>>>请输入聊天对象,按q退出<<<<<<")
		fmt.Scanln(&remoteName)
		if remoteName == "q" {
			return
		}
		if remoteName == "" {
			continue
		}
		client.chat_by_name(remoteName, true)
	}
}

func (client *Client) chat_by_name(remoteName string, isPrivate bool) {
	chatMsg := ""
	for {
		fmt.Println(">>>>>>请输入聊天内容,按q退出<<<<<<")
		fmt.Scanln(&chatMsg)
		if chatMsg == "q" {
			return
		}
		if len(chatMsg) == 0 {
			continue
		}
		var sendMsg string
		if isPrivate {
			sendMsg = "to?" + remoteName + "?" + chatMsg + "\n"
		} else {
			sendMsg = chatMsg + "\n"
		}
		client.sendMsgToServer(sendMsg)
	}
}

func (client *Client) updateName() {
	fmt.Println(">>>>>>请输入用户名<<<<<<")
	fmt.Scanln(&client.Name)
	sendMsg := "rename?" + client.Name + "\n"
	client.sendMsgToServer(sendMsg)
}

func (client *Client) sendMsgToServer(sendMsg string) {
	_, err := client.Conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
}
func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.公共频道")
	fmt.Println("2.个人频道")
	fmt.Println("3.修改名称")
	fmt.Println("0.退出")
	fmt.Scanln(&flag)
	fmt.Printf("flag: %v\n", flag)
	if flag >= 0 && flag <= 3 {
		client.Flag = flag
		return true
	} else {
		fmt.Println(">>>>>请输入合法范围内的数字<<<<<<")
		return false
	}
}

func (client *Client) Run() {
	for client.Flag != 0 {
		for !client.menu() {
		}
		switch client.Flag {
		case 1:
			// fmt.Println("公共频道选择...")
			client.chat_by_name("", false)
			break
		case 2:
			// fmt.Println("个人频道选择...")
			client.privateChat()
			break
		case 3:
			// fmt.Println("修改名称选择...")
			client.updateName()
			break
		}
	}
}
