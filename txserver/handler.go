package main

import (
	"context"
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

var handlerMap = map[string]func(*context.Context, *Command){
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

func add(ctx *context.Context, command *Command) {

}

func commit_buy(ctx *context.Context, command *Command) {

}

func cancel_buy(ctx *context.Context, command *Command) {

}

func commit_sell(ctx *context.Context, command *Command) {

}

func cancel_sell(ctx *context.Context, command *Command) {

}

func display_summary(ctx *context.Context, command *Command) {

}

func buy(ctx *context.Context, command *Command) {

}

func sell(ctx *context.Context, command *Command) {

}

func set_buy_amount(ctx *context.Context, command *Command) {

}

func set_buy_trigger(ctx *context.Context, command *Command) {

}

func set_sell_amount(ctx *context.Context, command *Command) {

}

func set_sell_trigger(ctx *context.Context, command *Command) {

}

func quote(ctx *context.Context, command *Command) {

}

func cancel_set_buy(ctx *context.Context, command *Command) {

}

func cancel_set_sell(ctx *context.Context, command *Command) {

}

func dumplog(ctx *context.Context, command *Command) {
	eventCollection := mongoClient.Database("test").Collection("events")
	if command.Username != "" {
		// get events for specified user
	} else {
		// get all events
	}



}

func handle(ctx *context.Context, command *Command) {
	handlerMap[command.Command](ctx, command)

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
