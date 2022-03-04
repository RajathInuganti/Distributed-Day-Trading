package main

import (
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

const (
	CONN_URL      = "192.168.4.2:4444"
	characters    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+=/%#"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func quote_server_connect() net.Conn {
	conn, err := net.Dial("tcp", CONN_URL)
	if err != nil {
		log.Printf("error while dialing tcp: %s\n", err)
		return nil
	}

	return conn
}

func parseQuote(arr []string) (price float64, timestamp int64, crypto string, err error) {
	crypto = arr[4]
	timestamp32, err := strconv.Atoi(arr[3])
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
func get_quote(stock string, username string) []string {

	log.Printf("get_quote() ran\n")

	var conn net.Conn

	conn = quote_server_connect()
	for conn == nil {
		conn = quote_server_connect()
	}

	log.Printf("Connected to quote server\n")

	_, err := conn.Write([]byte(stock + username + "\n"))
	if err != nil {
		return get_quote(stock, username)
	}

	log.Printf("Sent request to quote Server\n")

	result, err := ioutil.ReadAll(conn)
	if err != nil || result == nil {
		return get_quote(stock, username)
	}

	log.Printf("Got response from quote server : %s\n", string(result))

	err = conn.Close()
	if err != nil {
		log.Fatalf("Unable to close connection with UVic Quote server")
	}

	return strings.Split(string(result), ",")

}

func getFakeQuote() (price float64, timestamp int64, crypto string) {
	return rand.Float64() * 100, time.Now().Unix(), generateCryptoKey(44)
}

// solution taken from https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func generateCryptoKey(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(characters) {
			b[i] = characters[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
