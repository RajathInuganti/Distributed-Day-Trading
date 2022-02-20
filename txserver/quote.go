package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
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

//Use for testing on UVic machine
func get_qoute(username string, stock string) string {

	var conn net.Conn

	conn = quote_server_connect()
	for conn == nil {
		conn = quote_server_connect()
	}

	_, err := conn.Write([]byte(stock + username))
	if err != nil {
		return get_qoute(username, stock)
	}

	result, err := ioutil.ReadAll(conn)
	if err != nil || result == nil {
		return get_qoute(username, stock)
	}

	fmt.Println(string(result))

	err = conn.Close()
	if err != nil {
		log.Fatalf("Unable to close connection with UVic Quote server")
	}

	return string(result)

}
