package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	clients = make(map[*websocket.Conn]bool)
	mutex   sync.Mutex
)

func broadcast(message []byte) {
	mutex.Lock()
	defer mutex.Unlock()

	for client := range clients {
		err := client.WriteMessage(
			websocket.TextMessage,
			message,
		)

		if err != nil {
			fmt.Println("Ошибка отправки:", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Ошибка WebSocket:", err)
		return
	}

	mutex.Lock()
	clients[conn] = true
	mutex.Unlock()

	fmt.Println("Новый клиент подключился")

	defer func() {
		mutex.Lock()
		delete(clients, conn)
		mutex.Unlock()

		conn.Close()

		fmt.Println("Клиент отключился")
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		fmt.Println("Сообщение:", string(message))

		broadcast(message)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "C:/myfiles/LABS/3/PP/7/6/index.html")
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/ws", handleWebSocket)

	fmt.Println("Чат запущен: http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}