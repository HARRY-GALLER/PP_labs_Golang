package main

import (
	"bufio"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

func saveKeys(priv *rsa.PrivateKey) error {
	if err := os.MkdirAll("keys", 0700); err != nil {
		return err
	}
	privateBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}
	publicBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join("keys", "private.pem"), pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateBytes}), 0600); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join("keys", "public.pem"), pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicBytes}), 0644)
}

func loadPublic(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("неверный PEM")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("ожидался RSA-ключ")
	}
	return rsaKey, nil
}

func loadPrivate(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("неверный PEM")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("ожидался RSA-ключ")
	}
	return rsaKey, nil
}

func main() {
	aliceKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("Ошибка создания ключа:", err)
		return
	}
	if err := saveKeys(aliceKey); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Ключи сохранены в keys/private.pem и keys/public.pem")
	in := bufio.NewScanner(os.Stdin)
	fmt.Print("Алиса: введите сообщение для Боба: ")
	if !in.Scan() {
		return
	}
	message := in.Text()

	// Алиса использует закрытый ключ из файла.
	privateKey, err := loadPrivate("keys/private.pem")
	if err != nil {
		fmt.Println(err)
		return
	}
	digest := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Алиса передала сообщение:", message)
	fmt.Println("Алиса передала подпись (Base64):", base64.StdEncoding.EncodeToString(signature))

	// Боб получает сообщение/подпись и использует открытый ключ из файла.
	publicKey, err := loadPublic("keys/public.pem")
	if err != nil {
		fmt.Println(err)
		return
	}
	receivedDigest := sha256.Sum256([]byte(message))
	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, receivedDigest[:], signature); err != nil {
		fmt.Println("Боб: подпись неверна")
	} else {
		fmt.Println("Боб: подпись верна, сообщение не изменено")
	}
}
