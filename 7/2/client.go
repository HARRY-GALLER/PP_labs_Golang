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
		fmt.Println("Ошибка подключения:", err)
		return
	}
	defer conn.Close()

	fmt.Print("Введите сообщение: ")

	inputReader := bufio.NewReader(os.Stdin)

	message, err := inputReader.ReadString('\n')
	if err != nil {
		fmt.Println("Ошибка ввода:", err)
		return
	}

	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Ошибка отправки:", err)
		return
	}

	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Ошибка получения ответа:", err)
		return
	}

	fmt.Println("Ответ сервера:", strings.TrimSpace(response))
}