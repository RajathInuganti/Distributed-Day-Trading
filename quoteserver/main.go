package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
)

var src = rand.NewSource(time.Now().UnixNano())

const (
	characters    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+=/%#"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

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

	return string(b)
}

func sendQuote(conn net.Conn) {
	price, timestamp, crypto := getFakeQuote()
	responseString := fmt.Sprintf("%f,%d,%s", price, timestamp, crypto)

	conn.Write([]byte(responseString))

	defer conn.Close()
}

func main() {
	server, err := net.Listen("tcp", ":4444")
	if err != nil {
		panic(err)
	}

	log.Println("Listening on localhost:2000")

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go sendQuote(conn)
	}

}
