package main

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

func hashText(algorithm, text string) (string, error) {
	switch algorithm {
	case "1":
		sum := md5.Sum([]byte(text))
		return hex.EncodeToString(sum[:]), nil
	case "2":
		sum := sha256.Sum256([]byte(text))
		return hex.EncodeToString(sum[:]), nil
	case "3":
		sum := sha512.Sum512([]byte(text))
		return hex.EncodeToString(sum[:]), nil
	default:
		return "", fmt.Errorf("неизвестный алгоритм")
	}
}

func main() {
	in := bufio.NewScanner(os.Stdin)
	fmt.Println("Алгоритм: 1 — MD5, 2 — SHA-256, 3 — SHA-512")
	fmt.Print("Выберите алгоритм: ")
	if !in.Scan() {
		return
	}
	algorithm := strings.TrimSpace(in.Text())
	fmt.Print("Введите строку: ")
	if !in.Scan() {
		return
	}
	message := in.Text()
	value, err := hashText(algorithm, message)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Println("Хэш:", value)
	fmt.Print("Проверить целостность? (да/нет): ")
	if !in.Scan() || strings.ToLower(strings.TrimSpace(in.Text())) != "да" {
		return
	}
	fmt.Print("Введите строку для проверки: ")
	if !in.Scan() {
		return
	}
	checkText := in.Text()
	fmt.Print("Введите её хэш: ")
	if !in.Scan() {
		return
	}
	entered := strings.TrimSpace(in.Text())
	given, err := hex.DecodeString(entered)
	if err != nil {
		fmt.Println("Хэш неверен: неправильный формат hex")
		return
	}
	checkHex, _ := hashText(algorithm, checkText)
	expected, _ := hex.DecodeString(checkHex)
	if subtle.ConstantTimeCompare(given, expected) == 1 {
		fmt.Println("Целостность подтверждена: хэш совпадает")
	} else {
		fmt.Println("Целостность нарушена: хэш не совпадает")
	}
}
