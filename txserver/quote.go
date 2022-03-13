package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	CONN_URL = "quoteserver:4444"
)

func quote_server_connect() net.Conn {
	conn, err := net.Dial("tcp", CONN_URL)
	if err != nil {
		log.Printf("error while dialing tcp: %s\n", err)
		return nil
	}

	return conn
}

func parseQuote(arr []string) (price float64, timestamp int64, crypto string, err error) {
	crypto = arr[2]
	timestamp32, err := strconv.Atoi(arr[1])
	if err != nil {
		return 0, 0, "", err
	}

	timestamp = int64(timestamp32)

	price, err = strconv.ParseFloat(arr[0], 64)
	if err != nil {
		return 0, 0, "", nil
	}

	return price, timestamp, crypto, nil
}

//Use for testing on UVic machine
func get_quote(stock string, username string) ([]string, error) {

	log.Printf("get_quote() ran\n")

	var conn net.Conn

	conn = quote_server_connect()
	for conn == nil {
		conn = quote_server_connect()
	}

	log.Printf("Connected to quote server\n")

	_, err := conn.Write([]byte(fmt.Sprintf("%s,%s\n", stock, username)))
	for err != nil {
		_, err = conn.Write([]byte(fmt.Sprintf("%s,%s\n", stock, username)))
	}

	log.Printf("Sent request to quote Server\n")

	result := make([]byte, 1024)
	_, err = conn.Read(result)
	log.Printf("tried to read %s, error : %s\n", result, err)
	if err != nil {
		return nil, err
	}

	log.Printf("Got response from quote server : %s\n", string(result))

	err = conn.Close()
	if err != nil {
		log.Fatalf("Unable to close connection with UVic Quote server")
		return nil, err
	}

	return strings.Split(string(result), ","), nil

}
