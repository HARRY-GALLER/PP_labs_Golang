package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
)

func encrypt(text, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	} // AES: ключ 16, 24 или 32 байта
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	result := aead.Seal(nonce, nonce, []byte(text), nil)
	return base64.StdEncoding.EncodeToString(result), nil
}

func decrypt(encoded, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	if len(data) < aead.NonceSize() {
		return "", errors.New("шифртекст слишком короткий")
	}
	nonce, ciphertext := data[:aead.NonceSize()], data[aead.NonceSize():]
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func main() {
	in := bufio.NewScanner(os.Stdin)
	fmt.Print("Введите строку: ")
	if !in.Scan() {
		return
	}
	text := in.Text()
	fmt.Print("Введите секретный ключ (16/24/32 байта): ")
	if !in.Scan() {
		return
	}
	secret := in.Text()
	result, err := encrypt(text, secret)
	if err != nil {
		fmt.Println("Ошибка шифрования:", err)
		return
	}
	fmt.Println("Зашифрованная строка (Base64):", result)
	fmt.Print("Введите ключ для расшифровки: ")
	if !in.Scan() {
		return
	}
	givenKey := in.Text()
	original, err := decrypt(result, givenKey)
	if err != nil {
		fmt.Println("Ошибка расшифровки: ключ неверен или данные повреждены")
		return
	}
	fmt.Println("Расшифрованная строка:", original)
}
