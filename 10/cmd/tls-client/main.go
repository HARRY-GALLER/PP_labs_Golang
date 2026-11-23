package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
)

func main() {
	pair, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client-key.pem")
	if err != nil {
		log.Fatal(err)
	}
	caPEM, err := os.ReadFile("certs/ca.pem")
	if err != nil {
		log.Fatal(err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(caPEM) {
		log.Fatal("не удалось прочесть корневой сертификат")
	}
	conn, err := tls.Dial("tcp", "127.0.0.1:8443", &tls.Config{
		RootCAs: roots, ServerName: "localhost", Certificates: []tls.Certificate{pair}, MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		log.Fatal("сертификат сервера или TLS-соединение не прошли проверку: ", err)
	}
	defer conn.Close()
	fmt.Println("Защищённое соединение установлено, сертификат сервера проверен")
	in := bufio.NewScanner(os.Stdin)
	response := bufio.NewReader(conn)
	fmt.Print("Сообщение серверу: ")
	if !in.Scan() {
		return
	}
	if _, err := fmt.Fprintln(conn, in.Text()); err != nil {
		log.Fatal(err)
	}
	line, err := response.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(line)
}
