package main

import (
	"io"
	"log"
	"net"
	"strconv"
)

var counter = 0

func sendQuote(conn net.Conn) {
	request := make([]byte, 1024)

	for {
		numberOfBytes, err := conn.Read(request)
		if err != nil {
			if err == io.EOF {
				_ = conn.Close()
				log.Printf("connection closed")
				return
			}

			log.Printf("error while reading: %+v\n", err)
		}

		if numberOfBytes == 0 {
			continue
		}

		counter += 1

		log.Println(len(string(request)))
		// #nosec
		conn.Write([]byte(strconv.Itoa(counter) + "\n"))
	}
}

func main() {
	// #nosec
	server, err := net.Listen("tcp", ":4444")
	if err != nil {
		panic(err)
	}

	log.Println("Listening on localhost:4444")

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Printf("error while accepting: %+v\n", err)
			panic(err)
		}

		go sendQuote(conn)
	}
}
