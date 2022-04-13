package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// wg is used to wait for the go routine that receives data from the server
var wg sync.WaitGroup

// keeps track of the number of responses to be received from the server
var counter uint64

// set to true only when all requests have been sent to the server. This is used to stop the go routine that receives responses from the server
var allRequestsSent bool = false

// for storing the response bytes
var buffer []byte = []byte{}

// Command struct is a representation of an isolated command executed by a user
type Command struct {
	Command  string `json:"Command"`
	Username string `json:"Username"`
	Amount   string `json:"Amount"`
	Stock    string `json:"Stock"`
	Filename string `json:"Filename"`
}

type Response struct {
	Command string `json:"command"`
	Data    []byte `json:"data"`
	Error   string `json:"error"`
}

// FromStringToCommandStruct takes a line from the user command file as an input and returns a defined golang structure
func FromStringToCommandStruct(line string) (*Command, error) {
	/*
		The line variable value should have a similar format to this: '[1] ADD,oY01WVirLr,63511.53'
	*/
	line = strings.Split(line, " ")[1]
	commandVars := strings.Split(line, ",")
	cmd := commandVars[0]

	if cmd == "ADD" {
		return &Command{Command: cmd, Username: commandVars[1], Amount: commandVars[2]}, nil
	}

	if cmd == "COMMIT_BUY" || cmd == "CANCEL_BUY" || cmd == "COMMIT_SELL" || cmd == "CANCEL_SELL" || cmd == "DISPLAY_SUMMARY" {
		return &Command{Command: cmd, Username: commandVars[1]}, nil
	}

	if cmd == "BUY" || cmd == "SELL" || cmd == "SET_BUY_AMOUNT" || cmd == "SET_BUY_TRIGGER" || cmd == "SET_SELL_AMOUNT" || cmd == "SET_SELL_TRIGGER" {
		return &Command{Command: cmd, Username: commandVars[1], Stock: commandVars[2], Amount: commandVars[3]}, nil
	}

	if cmd == "QUOTE" || cmd == "CANCEL_SET_BUY" || cmd == "CANCEL_SET_SELL" {
		return &Command{Command: cmd, Username: commandVars[1], Stock: commandVars[2]}, nil
	}

	if cmd == "DUMPLOG" {
		if len(commandVars) == 3 {
			// case: DUMPLOG,userid,filename
			return &Command{Command: cmd, Username: commandVars[1], Filename: commandVars[2]}, nil
		} else {
			// case: DUMPLOG,filename
			return &Command{Command: cmd, Filename: commandVars[1]}, nil
		}
	}

	if cmd == "DISPLAY_SUMMARY" {
		return &Command{Command: cmd, Username: commandVars[1]}, nil
	}

	return nil, fmt.Errorf("unable to conver given line: %s into golang struct", line)
}

func checkError(e error, additionalMessage string) {
	if e != nil {
		log.Printf(additionalMessage+": %s\n", e)
		panic(e)
	}
}

func HandleCommand(command *Command, conn net.Conn) error {
	var buffer bytes.Buffer
	err := json.NewEncoder(&buffer).Encode(command)
	if err != nil {
		log.Printf("Error while encoding command: %+v", command)
		return err
	}

	payloadLengthInBinary := make([]byte, 8)
	binary.LittleEndian.PutUint64(payloadLengthInBinary, uint64(buffer.Len()))
	_, err = conn.Write(payloadLengthInBinary)
	for err != nil {
		_, err = conn.Write(payloadLengthInBinary)
	}

	_, err = conn.Write(buffer.Bytes())
	if err != nil {
		log.Printf("Error while writing command: %+v", command)
		return err
	}

	return nil
}

func HandleResponse(res *Response) error {
	if res.Error != "" {
		log.Printf("command: %s, Error: %s\n", res.Command, res.Error)
		return nil
	}

	if res.Command == "DUMPLOG" {
		file, err := os.Create("logfile.xml")
		if err != nil {
			log.Printf("error while creating file: %s\n", err)
			return err
		}

		_, err = file.Write(res.Data)
		if err != nil {
			log.Printf("Error while writing response body to file: %s\n", err)
			return err
		}

		err = file.Close()
		if err != nil {
			log.Printf("Error while closing file: %s\n", err)
		}

		log.Printf("Contents successfully written to logfile.xml\n")
		return nil
	} else {
		log.Printf("%s\n", res.Data)
	}

	return nil
}

func processMessage(msg []byte, conn net.Conn) {
	response := &Response{}
	err := json.Unmarshal(msg, response)
	if err != nil {
		log.Printf("Error while unmarshalling response: %s, error: %s\n", string(msg), err)
	}

	err = HandleResponse(response)
	if err != nil {
		log.Printf("Error while handling response: %s, error: %s\n", string(msg), err)
	}

	atomic.AddUint64(&counter, ^uint64(0))
	if allRequestsSent && atomic.LoadUint64(&counter) == 0 {
		log.Printf("all requests sent and responses received, closing connection..\n")
		defer conn.Close()
		defer wg.Done()
		return
	}
}

func ReadResponse(conn net.Conn) {
	response := make([]byte, 1024*10)

	for {
		numberOfBytes, err := conn.Read(response)
		if err != nil {
			if err == io.EOF {
				continue
			}

			log.Printf("error while reading: %+v\n", err)
		}

		if numberOfBytes == 0 {
			continue
		}

		openBracketIndex := -1
		closeBracketIndex := -1
		messageFound := false
		for i, b := range response[:numberOfBytes] {
			if b == '{' {
				openBracketIndex = i
			}
			if b == '}' {
				closeBracketIndex = i
			}

			if openBracketIndex != -1 && closeBracketIndex != -1 && closeBracketIndex > openBracketIndex {
				processMessage(response[openBracketIndex:closeBracketIndex+1], conn)
				openBracketIndex = -1
				closeBracketIndex = -1
				messageFound = true
			}
		}

		if openBracketIndex != -1 && closeBracketIndex == -1 {
			buffer = append([]byte{}, response[openBracketIndex:numberOfBytes]...)
		} else if openBracketIndex == -1 && closeBracketIndex != -1 {
			msg := append(buffer[:], response[:closeBracketIndex+1]...)
			processMessage(msg, conn)
			buffer = []byte{}
		} else if openBracketIndex != -1 && closeBracketIndex != -1 && closeBracketIndex < openBracketIndex {
			msg := append(buffer[:], response[:closeBracketIndex+1]...)
			processMessage(msg, conn)
			buffer = append([]byte{}, response[closeBracketIndex+1:]...)
		} else {
			// do nothing
		}

		if openBracketIndex == -1 && closeBracketIndex == -1 && !messageFound {
			buffer = append(buffer[:], response[:numberOfBytes]...)
		}

	}
}

func MakeSocketConnection() net.Conn {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Printf("Error while dialing: %s\n", err)
		panic(err)
	}

	return conn
}

func main() {

	s1 := time.Now()
	conn := MakeSocketConnection()

	if len(os.Args) != 2 {
		fmt.Println("Please follow the following format: go run res.Command.go <path_to_workload_file.txt>")
		panic("Unexpected number of arguments")
	}

	commandsFilePath := os.Args[1]
	data, err := os.ReadFile(filepath.Clean(commandsFilePath))
	checkError(err, "Error while reading file")

	lines := strings.Split(string(data), "\n")

	go ReadResponse(conn)
	wg.Add(1)

	for _, line := range lines {
		if line == "" {
			continue
		}

		requestData, err := FromStringToCommandStruct(line)
		checkError(err, "Couldn't convert line from file to command struct")

		if err != nil {
			log.Fatal(err)
		}

		err = HandleCommand(requestData, conn)
		if err != nil {
			log.Printf("Error while handling command %+v: %s\n", requestData, err)
		}

		atomic.AddUint64(&counter, 1)
	}

	allRequestsSent = true
	log.Printf("All requests sent, waiting for responses..\n")

	wg.Wait()

	log.Printf("Took %s\n", time.Since(s1))
}
