package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	network     = "tcp"
	serverPort  = "localhost:8080"
	successResp = "OK\n"
	timeout     = 5 * time.Second
)

func main() {
	conn, err := net.DialTimeout(network, serverPort, timeout)
	if err != nil {
		fmt.Printf("Ошибка подключения к %s: %v\n", serverPort, err)
		os.Exit(1)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(timeout))

	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Ошибка чтения ответа от %s: %v\n", serverPort, err)
		os.Exit(1)
	}

	if response != successResp {
		fmt.Printf("Неверный ответ от сервера (ожидали %q, получили %q)\n",
			successResp, response)
		os.Exit(1)
	}

	fmt.Print(response)
}
