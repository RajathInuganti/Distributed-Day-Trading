package main

import (
	"fmt"
	"net"

	//"os"
	//"strings"
	"io/ioutil"
)

const (
	CONN_URL = "192.168.4.2:4444"
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

func quote_server_connect() net.Conn {

	conn, err := net.Dial("tcp", CONN_URL)
	if err != nil {
		return nil
	}

	return conn
}

//Use for testing on UVic machine
func get_qoute(username string, stock string) string {

	var conn net.Conn

	conn = quote_server_connect()
	for conn == nil {
		conn = quote_server_connect()
	}

	defer conn.Close()

	_, err := conn.Write([]byte(stock + username))
	if err != nil {
		return get_qoute(username, stock)
	}

	result, err := ioutil.ReadAll(conn)
	if err != nil || result == nil {
		get_qoute(username, stock)
	}

	fmt.Println(string(result))

	return string(result)

}
