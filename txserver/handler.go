package main

import (
	"fmt"
	"net"
	//"os"
	//"strings"
	"strconv"
	"io/ioutil"
)

const (
	CONN_HOST = "192.168.4.2"
	CONN_PORT = 4444
)


//Use for testing on UVic machine
func callQuery() {
	connection_string := CONN_HOST + strconv.Itoa(CONN_PORT)
	c, err := net.Dial("tcp", connection_string)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = c.Write([]byte("BKM Warlock"))

	result, err := ioutil.ReadAll(c)

	fmt.Println(string(result))

	
}

// func parseStruct() {

// }
