package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

var wg sync.WaitGroup

func handleClient(conn net.Conn) {
	defer wg.Done()
	defer conn.Close()

	fmt.Println("Клиент подключился:", conn.RemoteAddr())

	reader := bufio.NewReader(conn)

	message, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Ошибка чтения:", err)
		return
	}

	message = strings.TrimSpace(message)

	fmt.Println("Получено от", conn.RemoteAddr(), ":", message)

	_, err = conn.Write([]byte("Сообщение получено сервером\n"))
	if err != nil {
		fmt.Println("Ошибка отправки ответа:", err)
		return
	}

	fmt.Println("Клиент обработан:", conn.RemoteAddr())
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
		return
	}

	fmt.Println("Сервер запущен на порту 8080")
	fmt.Println("Для остановки нажмите Ctrl+C")

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	go func() {
		<-ctx.Done()

		fmt.Println("\nПолучен сигнал завершения")
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()

		if err != nil {
			select {
			case <-ctx.Done():
				fmt.Println("Сервер прекращает принимать подключения")
				wg.Wait()
				fmt.Println("Все соединения завершены")
				return
			default:
				fmt.Println("Ошибка подключения:", err)
				continue
			}
		}

		wg.Add(1)

		go handleClient(conn)
	}
}