package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

const CONN_URL = "localhost:4444"

func quote_server_connect() net.Conn {
	conn, err := net.Dial("tcp", CONN_URL)
	if err != nil {
		log.Printf("error while dialing tcp: %s\n", err)
		return nil
	}

	return conn
}

func main() {
	var conn net.Conn
	conn = quote_server_connect()
	for conn == nil {
		conn = quote_server_connect()
	}

	counter := 0
	for counter < 2 {
		_, err := conn.Write([]byte(fmt.Sprintf("%s,%s\n", "hello", "there")))
		for err != nil {
			_, err = conn.Write([]byte(fmt.Sprintf("%s,%s\n", "hello", "there")))
		}

		time.Sleep(time.Second)

		counter += 1
	}

	result := make([]byte, 1024)
	_, err := conn.Read(result)
	if err != nil {
		log.Printf("error while reading: %s\n", err)
		panic(err)
	}

	log.Printf("got response: %s\n", result)

	err = conn.Close()
	if err != nil {
		log.Printf("error while closing: %+v\n", err)
		panic(err)
	}
}
