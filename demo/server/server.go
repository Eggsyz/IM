package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

/**
 * @Author: eggsy
 * @Description:
 * @File:  server
 * @Version: 1.0.0
 * @Date: 2020-03-17 10:47
 */
// 保存客户端信息
type Client struct {
	Name string
	Addr string
	// 保存需要发送给client的信息
	Messages chan string
}

var onlineClientMap = make(map[string]Client)
var message = make(chan string)

// 格式化消息
func fmtMsg(addr, msg string) string {
	return fmt.Sprintf("[%s]: %s", addr, msg)
}

// 发送消息给client
func sendMsgToClient(client Client, conn net.Conn) error {
	for msg := range client.Messages {
		if _, err := conn.Write([]byte(msg)); err != nil {
			return fmt.Errorf("failed to send message. client: %s, err: %v", client.Addr, err)
		}
	}
	return nil
}

// 广播消息
func broadcast() {
	for {
		msg := <-message
		for _, client := range onlineClientMap {
			client.Messages <- msg
		}
	}
}

func handler(conn net.Conn) error {
	defer conn.Close()
	addr := conn.LocalAddr().String()
	client := Client{
		Addr:     addr,
		Name:     addr,
		Messages: make(chan string),
	}
	// 将客户端添加入上线客户端
	onlineClientMap[addr] = client
	// 发送消息给指定客户端
	go sendMsgToClient(client, conn)
	// 1. 广播登陆消息
	message <- fmtMsg(addr, "login\n")

	// 2. 处理用户消息
	var quit = make(chan bool)
	var chat = make(chan bool)
	var timeout = time.Minute * 5
	// 处理消息
	go func() {
		var buffer = make([]byte, 4096)
		for {
			n, err := conn.Read(buffer)
			if err != nil {
				fmt.Printf("failed to read message, %v\n", err)
				return
			}
			if n == 0 {
				fmt.Printf("%s offline \n", addr)
				quit <- true
				return
			}
			// 私发消息: 消息格式 pvt|addr|message
			strs := strings.Split(string(buffer[:n]), "|")
			if strs[0] == "pvt" {
				if client, ok := onlineClientMap[strs[1]]; ok {
					client.Messages <- fmtMsg(strs[1], strs[2])
					conn.Write([]byte("send success"))
				} else {
					conn.Write([]byte("no such people"))
				}
			} else { // 广播消息
				message <- fmtMsg(addr, string(buffer[:n]))
			}
			chat <- true
		}
	}()
	// 处理超时或者退出
	for {
		select {
		case <-chat:

		case <-time.After(timeout):
			delete(onlineClientMap, addr)
			close(client.Messages)
			message <- fmtMsg(addr, "timeout")
			return nil
		case <-quit:
			delete(onlineClientMap, addr)
			close(client.Messages)
			message <- fmtMsg(addr, "quit")
			return nil
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:9092")
	defer listener.Close()
	if err != nil {
		panic(err)
	}
	go broadcast()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handler(conn)
	}
}
