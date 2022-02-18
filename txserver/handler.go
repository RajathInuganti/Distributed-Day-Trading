package main

import (
	"fmt"
	"net"

	//"os"
	//"strings"
	"io/ioutil"
	"strconv"
)

const (
	CONN_HOST = "192.168.4.2"
	CONN_PORT = 4444
)

var handlerMap = map[string]func(*Command){
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

func add(command *Command) {

}

func commit_buy(command *Command) {

}

func cancel_buy(command *Command) {

}

func commit_sell(command *Command) {

}

func cancel_sell(command *Command) {

}

func display_summary(command *Command) {

}

func buy(command *Command) {

}

func sell(command *Command) {

}

func set_buy_amount(command *Command) {

}

func set_buy_trigger(command *Command) {

}

func set_sell_amount(command *Command) {

}

func set_sell_trigger(command *Command) {

}

func quote(command *Command) {

}

func cancel_set_buy(command *Command) {

}

func cancel_set_sell(command *Command) {

}

func dumplog(command *Command) {

}

func handle(command *Command) {

	handlerMap[command.Command](command)

}
