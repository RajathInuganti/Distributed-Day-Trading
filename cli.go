package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Command struct is a representation of an isolated command executed by a user
type Command struct {
	Command  string `json:"Command"`
	Username string `json:"Username"`
	Amount   string `json:"Amount"`
	Stock    string `json:"Stock"`
	Filename string `json:"Filename"`
}

type Response struct {
	Data  []byte `json:"data"`
	Error string `json:"error"`
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

func HandleCommand(command *Command) error {
	var buffer bytes.Buffer
	err := json.NewEncoder(&buffer).Encode(command)
	if err != nil {
		log.Printf("Error while encoding command: %+v", command)
		return err
	}

	res, err := http.Post("http://localhost:8080/", "application/json", &buffer)
	if err != nil {
		log.Printf("Error while sending request: %s for command: %+v", err, *command)
		return err
	}

	err = HandleResponse(command, res)
	if err != nil {
		log.Printf("Error while handling response for cmd: %+v: %s\n", command, err)
	}
	return nil
}

func HandleResponse(cmd *Command, res *http.Response) error {
	log.Printf("Got response: %s for %+v", res.Status, cmd)

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error while reading response body: %s\n", err)
		return err
	}

	log.Printf("Response body: %s\n", string(bodyBytes))

	responseStruct := &Response{}
	err = json.Unmarshal(bodyBytes, responseStruct)
	if err != nil {
		log.Printf("Error while unmarshalling response body: %s\n", err)
		return err
	}

	if responseStruct.Error != "" {
		log.Printf("Got an error in the response for command: %+v, error: %s\n", cmd, responseStruct.Error)
		return nil
	}

	if cmd.Command == "DUMPLOG" {
		err = ioutil.WriteFile(cmd.Filename, responseStruct.Data, 0444)
		if err != nil {
			log.Printf("Error while writing response body to file: %s\n", err)
			return err
		}
		fmt.Printf("Contents successfully written to %s\n", cmd.Filename)
		return nil
	}

	// For DISPLAY_SUMMARY or other commands
	fmt.Printf("%s\n", string(responseStruct.Data))
	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Please follow the following format: go run cmd.go <path_to_workload_file.txt>")
		panic("Unexpected number of arguments")
	}

	commandsFilePath := os.Args[1]
	data, err := os.ReadFile(filepath.Clean(commandsFilePath))
	checkError(err, "Error while reading file")

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}

		requestData, err := FromStringToCommandStruct(line)
		checkError(err, "Couldn't convert line from file to command struct")

		fmt.Printf("iteration: %d requestData: %#v\n", i+1, requestData)

		if err != nil {
			log.Fatal(err)
		}

		err = HandleCommand(requestData)
		if err != nil {
			log.Printf("Error while handling command %+v: %s\n", requestData, err)
		}
	}

}
