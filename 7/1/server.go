package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
		return
	}
	defer listener.Close()

	fmt.Println("TCP-сервер запущен на порту 8080")

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Ошибка подключения:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Клиент подключился:", conn.RemoteAddr())

	message, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Ошибка чтения:", err)
		return
	}

	fmt.Println("Получено сообщение:", message)

	_, err = conn.Write([]byte("Сообщение успешно получено\n"))
	if err != nil {
		fmt.Println("Ошибка отправки ответа:", err)
		return
	}

	fmt.Println("Ответ отправлен клиенту")
}