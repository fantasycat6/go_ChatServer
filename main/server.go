package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	clients      = make(map[int]net.Conn) // 存储在线客户端
	availableIDs = []int{}                // 存储可用的客户端编号
	mu           sync.Mutex               // 用于保护共享资源的锁
)

func handleConnection(conn net.Conn, id int) {
	defer conn.Close()
	fmt.Printf("客户端 %d 已连接：%s\n", id, conn.RemoteAddr().String())

	mu.Lock()
	clients[id] = conn
	mu.Unlock()

	// 通知客户端它的编号
	fmt.Fprintf(conn, "您已连接到服务器，您的客户端编号是 %d\n", id)

	// 通知所有客户端当前在线客户端列表
	broadcastClientList()

	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("客户端 %d 断开连接\n", id)
			mu.Lock()
			delete(clients, id)
			availableIDs = append(availableIDs, id) // 将编号放回可用编号池
			broadcastClientList()
			broadcastDisconnection(id)
			mu.Unlock()
			return
		}

		message = strings.TrimSpace(message)
		if strings.HasPrefix(message, "@") {
			targetID, msg := parseDirectMessage(message)
			mu.Lock()
			if targetConn, exists := clients[targetID]; exists {
				fmt.Fprintf(targetConn, "客户端 %d 私聊：%s\n", id, msg)
			} else {
				fmt.Fprintf(conn, "客户端 %d 不在线或不存在\n", targetID)
			}
			mu.Unlock()
		} else {
			broadcastMessage(id, message)
		}
	}
}

func getNewClientID() int {
	if len(availableIDs) > 0 {
		id := availableIDs[0]
		availableIDs = availableIDs[1:]
		return id
	}
	if len(clients) == 0 {
		return 1
	}
	maxID := 0
	for id := range clients {
		if id > maxID {
			maxID = id
		}
	}
	return maxID + 1
}

func broadcastMessage(senderID int, message string) {
	for id, client := range clients {
		if id != senderID {
			fmt.Fprintf(client, "客户端 %d：%s\n", senderID, message)
		}
	}
}

func broadcastClientList() {
	clientList := "当前在线客户端："
	for id := range clients {
		clientList += fmt.Sprintf(" %d", id)
	}
	clientList += "\n"
	for _, client := range clients {
		fmt.Fprint(client, clientList)
	}
}

func broadcastDisconnection(disconnectedID int) {
	message := fmt.Sprintf("客户端 %d 已断开连接\n", disconnectedID)
	for _, client := range clients {
		fmt.Fprint(client, message)
	}
}

func parseDirectMessage(message string) (int, string) {
	parts := strings.SplitN(message, " ", 2)
	targetID, _ := strconv.Atoi(parts[0][1:]) // 提取目标客户端ID
	msg := ""
	if len(parts) > 1 {
		msg = parts[1]
	}
	return targetID, msg
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("错误：", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("服务器已启动，在端口 8080 监听...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("连接错误：", err.Error())
			continue
		}

		mu.Lock()
		clientID := getNewClientID() // 获取新的客户端编号
		mu.Unlock()

		go handleConnection(conn, clientID)
	}
}
