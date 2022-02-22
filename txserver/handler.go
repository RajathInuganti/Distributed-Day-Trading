package main

import "log"

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

	responseData, err := handlerMap[command.Command](command)
	if err != nil {
		log.Println(err)
		return &Response{Error: err.Error()}
	}

	return &Response{Data: responseData}
}


