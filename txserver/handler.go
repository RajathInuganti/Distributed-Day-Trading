package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
)

//"os"
//"strings"

var handlerMap = map[string]func(*Command)([]byte, error){
	"ADD":              add,
	"COMMIT_BUY":       commit_buy,
	"CANCEL_BUY":       cancel_buy,
	"COMMIT_SELL":      commit_sell,
	"CANCEL_SELL":      cancel_sell,
	"DISPLAY_SUMMARY":  display_summary,
	"BUY":              buy,
	"SELL":             sell,
	"SET_BUY_AMOUNT":   set_buy_amount,
	"SET_BUY_TRIGGER":  set_buy_trigger,
	"SET_SELL_AMOUNT":  set_sell_amount,
	"SET_SELL_TRIGGER": set_sell_trigger,
	"QUOTE":            quote,
	"CANCEL_SET_BUY":   cancel_set_buy,
	"CANCEL_SET_SELL":  cancel_set_sell,
	"DUMPLOG":          dumplog,
}

func add(command *Command) ([]byte, error) {
	return []byte{}, nil
}

func commit_buy(command *Command) ([]byte, error) {
	return []byte{}, nil
}

func cancel_buy(command *Command) ([]byte, error) {
	return []byte{}, nil
}

func commit_sell(command *Command) ([]byte, error) {
	return []byte{}, nil
}

func cancel_sell(command *Command) ([]byte, error) {
	return []byte{}, nil
}

func display_summary(command *Command) ([]byte, error) {
	return []byte{}, nil
}

func buy(command *Command) ([]byte, error) {
	return []byte{}, nil
}

func sell(command *Command) ([]byte, error) {
	return []byte{}, nil
}

func set_buy_amount(command *Command) ([]byte, error) {
	return []byte{}, nil
}

func set_buy_trigger(command *Command) ([]byte, error) {
	return []byte{}, nil
}

func set_sell_amount(command *Command) ([]byte, error) {
	return []byte{}, nil
}

func set_sell_trigger(command *Command) ([]byte, error) {
	return []byte{}, nil
}

func quote(command *Command) ([]byte, error) {
	return []byte{}, nil
}

func cancel_set_buy(command *Command) ([]byte, error) {
	return []byte{}, nil
}

func cancel_set_sell(command *Command) ([]byte, error) {
	return []byte{}, nil
}

func dumplog(command *Command) ([]byte, error) {
	return []byte{}, nil
}

func handle(command *Command) *Response{
	err := VerifyAndParseRequestData(command)
	if err != nil {
		return &Response{ Error: err.Error() }
	}

	responseData, err := handlerMap[command.Command](command)
	if err != nil {
		log.Println(err)
		return &Response{Error: err.Error()}
	}

	return &Response{Data: responseData}
}


func VerifyAndParseRequestData(command *Command) error {
	if command.Command == "" {
		return errors.New("no command specified in the command message")
	}

	amountIntValue, AmountNotConvertibleToIntError := strconv.Atoi(fmt.Sprintf("%v",command.Amount))
	usernameEmpty := command.Username == ""
	stockSymbolEmpty := command.Stock == ""
		

	if command.Command == "ADD" && (command.Username == "" || AmountNotConvertibleToIntError != nil){
		return errors.New("either the provided username is empty or the amount specified is not a number")
	}

	if (command.Command == "COMMIT_BUY" || command.Command == "CANCEL_BUY" || command.Command == "COMMIT_SELL" ||
		command.Command == "CANCEL_SELL" || command.Command == "DISPLAY_SUMMARY") && (usernameEmpty) {
		return errors.New("an empty Username was specified")
	}

	if (command.Command == "BUY" || command.Command == "SELL" || command.Command == "SET_BUY_AMOUNT" ||
	 	command.Command == "SET_BUY_TRIGGER" || command.Command == "SET_SELL_AMOUNT" || command.Command == "SET_SELL_TRIGGER") && 
		(usernameEmpty || stockSymbolEmpty || AmountNotConvertibleToIntError != nil) {
		return errors.New("either the provided username is empty or the amount specified is not a number or a stockSymbol was not specified")
	}

	if command.Command == "QUOTE" || command.Command == "CANCEL_SET_BUY" || command.Command == "CANCEL_SET_SELL" &&
		(usernameEmpty || stockSymbolEmpty) {
		return errors.New("either the provided username is empty or the stockSymbol was not specified")
	}

	if command.Command == "DISPLAY_SUMMARY" && usernameEmpty {
		return errors.New("username cannot be empty")
	}

	// setting Amount to an integer value so that it can be used in the rest of the code
	if AmountNotConvertibleToIntError == nil {
		command.Amount = amountIntValue
	}

	return nil
}

