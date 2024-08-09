package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("连接错误：", err)
		os.Exit(1)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("已连接到服务器。您可以：")
	fmt.Println("1. 输入消息并回车，进行广播聊天")
	fmt.Println("2. 使用 '@客户端编号 消息' 的格式私聊，例如：@2 你好")

	go func() {
		serverReader := bufio.NewReader(conn)
		for {
			message, err := serverReader.ReadString('\n')
			if err != nil {
				fmt.Println("\n与服务器连接断开")
				os.Exit(0)
			}

			// 添加消息类型提示
			if strings.Contains(message, "私聊") {
				fmt.Print("\n[私聊接收] ➤ ", message)
			} else if strings.Contains(message, "已连接") || strings.Contains(message, "断开连接") || strings.Contains(message, "当前在线客户端") {
				fmt.Print("\n[服务器消息] ➤ ", message)
			} else {
				fmt.Print("\n[广播接收] ➤ ", message)
			}

			// 重新显示发送提示符
			fmt.Print("[我] ➤ ")
		}
	}()

	for {
		fmt.Print("[我] ➤ ")
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)
		if message != "" {
			fmt.Fprintf(conn, message+"\n")
		}
	}
}
