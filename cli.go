package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Command struct is a representation of an isolated command executed by a user
type Command struct {
	Command string `json:"Command"`
	Username string `json:"Username"`
	Amount string `json:"Amount"`
	Stock string `json:"Stock"`
	Filename string `json:"Filename"`
}

// FromStringToCommandStruct takes a line from the user command file as an input and returns a defined golang structure 
func FromStringToCommandStruct(line string) *Command {
	/* 
		The line variable value should have a similar format to this: '[1] ADD,oY01WVirLr,63511.53' 
	*/
	line = strings.Split(line, " ")[1]
	commandVars := strings.Split(line, ",")
	cmd := commandVars[0]

	if cmd == "ADD" {
		return &Command{Command: cmd, Username: commandVars[1], Amount: commandVars[2]}
	} else if (cmd == "COMMIT_BUY" || cmd == "CANCEL_BUY" || cmd == "COMMIT_SELL" || cmd == "CANCEL_SELL" || cmd == "DISPLAY_SUMMARY") {
		return &Command{Command: cmd, Username: commandVars[1]}
	} else if (cmd == "BUY" || cmd == "SELL" || cmd == "SET_BUY_AMOUNT" || cmd == "SET_BUY_TRIGGER" || cmd == "SET_SELL_AMOUNT" || cmd == "SET_SELL_TRIGGER") {
		return &Command{Command: cmd, Username: commandVars[1], Stock: commandVars[2], Amount: commandVars[3]}
	} else if (cmd == "QUOTE" || cmd == "CANCEL_SET_BUY" || cmd == "CANCEL_SET_SELL") {
		return &Command{Command: cmd, Username: commandVars[1], Stock: commandVars[2]}
	} else if (cmd == "DUMPLOG") {
		if len(commandVars) == 3 {
			// case: DUMPLOG,userid,filename
			return &Command{Command: cmd, Username: commandVars[1], Filename: commandVars[2]}
		} else {
			// case: DUMPLOG,filename
			return &Command{Command: cmd, Filename: commandVars[1]}
		}
	} else if (cmd == "DISPLAY_SUMMARY") {
		return &Command{Command: cmd, Username: commandVars[1]}
	} else {
		fmt.Printf("Command received: %s, line: %s\n", cmd, line)
		panic("Unknown command received")
	}

}


func checkError(e error, additionalMessage string) {
	if e != nil {
		fmt.Println(additionalMessage)
		panic(e)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Please follow the following format: go run cmd.go <path_to_workload_file.txt>")
		panic("Unexpected number of arguments")
	}

	commandsFilePath := os.Args[1]
	data, err := os.ReadFile(filepath.Clean(commandsFilePath))
	checkError(err, "Error while reading")
	
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}

		requestData := FromStringToCommandStruct(line)
		fmt.Printf("iteration: %d requestData: %#v\n", i+1, requestData)

		// parsedJson, err := json.Marshal(requestData)
		// checkError(err, "Couldn't parse golang struct to JSON")

		// _, err = http.Post("http://localhost:8080/", "application/json", bytes.NewBuffer(parsedJson))
		// checkError(err, "Got error while doing a post request")
	}

}