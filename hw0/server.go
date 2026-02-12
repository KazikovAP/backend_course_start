package main

import (
	"log"
	"net"
)

const (
	defaultNetwork = "tcp"
	defaultPort    = ":8080"
	welcomeMessage = "OK\n"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	_, err := conn.Write([]byte(welcomeMessage))
	if err != nil {
		log.Println("Ошибка отправки:", err)
		return
	}
}

func main() {
	listener, err := net.Listen(defaultNetwork, defaultPort)
	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
	defer listener.Close()

	log.Printf("Сервер запущен на %s", defaultPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Ошибка подключения:", err)
			continue
		}

		go handleConnection(conn)
	}
}
