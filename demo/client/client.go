package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

/**
 * @Author: eggsy
 * @Description:
 * @File:  client
 * @Version: 1.0.0
 * @Date: 2020-03-17 21:53
 */

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:9092")
	defer conn.Close()
	if err != nil {
		panic(err)
	}
	response := bufio.NewReader(conn)
	// 开一个协程专门负责接收消息
	go func() {
		for {
			msg, _ := response.ReadString('\n')
			fmt.Println("Received:", string(msg))
		}
	}()
	for {
		reader := bufio.NewReader(os.Stdin)
		var buffer = make([]byte, 4096)
		n, _ := reader.Read(buffer)
		conn.Write(buffer[:n])
	}
}
