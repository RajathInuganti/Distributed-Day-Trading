package main

import (
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	CONN_URL = "192.168.4.2:4444"
)

func quote_server_connect() net.Conn {

	conn, err := net.Dial("tcp", CONN_URL)
	if err != nil {
		return nil
	}

	return conn
}

func parseQuote(arr []string) (price float64, timestamp int64, crypto string) {
	crypto = arr[4]
	timestamp32, _ := strconv.Atoi(arr[3])
	timestamp = int64(timestamp32)
	price, _ = strconv.ParseFloat(arr[0], 64)

	return price, timestamp, crypto
}

//Use for testing on UVic machine
func get_quote(stock string, username string) []string {

	var conn net.Conn

	conn = quote_server_connect()
	for conn == nil {
		conn = quote_server_connect()
	}

	_, err := conn.Write([]byte(stock + username + "\n"))
	if err != nil {
		return get_quote(stock, username)
	}

	result, err := ioutil.ReadAll(conn)
	if err != nil || result == nil {
		return get_quote(stock, username)
	}

	err = conn.Close()
	if err != nil {
		log.Fatalf("Unable to close connection with UVic Quote server")
	}

	return strings.Split(string(result), ",")

}
