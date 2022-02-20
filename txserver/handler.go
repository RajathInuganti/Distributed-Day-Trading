package main

import (
	"context"
	"fmt"
	"log"
	"net"

	//"os"
	//"strings"
	"io/ioutil"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	CONN_HOST = "192.168.4.2"
	CONN_PORT = 4444
)

var handlerMap = map[string]func(*context.Context, *Command) error {
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

func add(ctx *context.Context, command *Command) error {
	return nil
}

func commit_buy(ctx *context.Context, command *Command) error {
	return nil
}

func cancel_buy(ctx *context.Context, command *Command) error {
	return nil
}

func commit_sell(ctx *context.Context, command *Command) error {
	return nil
}

func cancel_sell(ctx *context.Context, command *Command) error {
	return nil
}

func display_summary(ctx *context.Context, command *Command) error {
	return nil
}

func buy(ctx *context.Context, command *Command) error {
	return nil
}

func sell(ctx *context.Context, command *Command) error {
	return nil
}

func set_buy_amount(ctx *context.Context, command *Command) error {
	return nil
}

func set_buy_trigger(ctx *context.Context, command *Command) error {
	return nil
}

func set_sell_amount(ctx *context.Context, command *Command) error {
	return nil
}

func set_sell_trigger(ctx *context.Context, command *Command) error {
	return nil
}

func quote(ctx *context.Context, command *Command) error {
	return nil
}

func cancel_set_buy(ctx *context.Context, command *Command) error {
	return nil
}

func cancel_set_sell(ctx *context.Context, command *Command) error {
	return nil
}

func dumplog(ctx *context.Context, command *Command) error {
	eventCollection := mongoClient.Database("test").Collection("Events")
	var cursor *mongo.Cursor
	var err error
	if command.Username != "" {
		// get events for specified user
		filter := bson.M{"user": command.Username}
		cursor, err = eventCollection.Find(*ctx, filter)
		if err != nil {
			log.Printf("Error getting events from the Events collectino for the user %s, query: %+v %s", command.Username, filter, err)
		}
	} else {
		// get all events
		cursor, err = eventCollection.Find(*ctx, bson.D{})
		if err != nil {
			log.Printf("Error getting all the events from the Events collection, error:  %s", err)
			return err
		}
	}

	defer cursor.Close(*ctx)

	return nil
}

func handle(ctx *context.Context, command *Command) {
	err := handlerMap[command.Command](ctx, command)
	if err != nil {
		log.Printf("Error handling command %+v, error: %s", command, err)
	}

	// we should retry if handling fails.
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
