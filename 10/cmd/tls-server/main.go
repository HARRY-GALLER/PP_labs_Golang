package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	pair, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server-key.pem")
	if err != nil {
		log.Fatal(err)
	}
	caPEM, err := os.ReadFile("certs/ca.pem")
	if err != nil {
		log.Fatal(err)
	}
	trusted := x509.NewCertPool()
	if !trusted.AppendCertsFromPEM(caPEM) {
		log.Fatal("не удалось прочесть корневой сертификат")
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{pair}, ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: trusted, MinVersion: tls.VersionTLS12}
	ln, err := tls.Listen("tcp", "127.0.0.1:8443", cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	fmt.Println("TLS-сервер запущен на 127.0.0.1:8443")
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go func(c io.ReadWriteCloser) {
			defer c.Close()
			tc := c.(*tls.Conn)
			if err := tc.Handshake(); err != nil {
				log.Println("TLS:", err)
				return
			}
			fmt.Println("Клиент с проверенным сертификатом подключился")
			reader := bufio.NewScanner(c)
			for reader.Scan() {
				message := reader.Text()
				fmt.Println("Получено:", message)
				fmt.Fprintln(c, "Сервер получил: "+message)
			}
		}(conn)
	}
}
