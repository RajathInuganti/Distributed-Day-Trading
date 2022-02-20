package main

import (
	"context"
	"day-trading/txserver/event"
	"encoding/xml"
	"log"

	//"os"
	//"strings"

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

	// fetch results from mongo
	var cursor *mongo.Cursor
	var err error
	if command.Username != "" {
		// get events for specified user
		filter := bson.M{"user": command.Username}
		cursor, err = eventCollection.Find(*ctx, filter)
		if err != nil {
			log.Printf("Error getting events from the Events collection for the user %s, query: %+v %s", command.Username, filter, err)
		}
	} else {
		// get all events
		cursor, err = eventCollection.Find(*ctx, bson.D{})
		if err != nil {
			log.Printf("Error getting all the events from the Events collection, error:  %s", err)
			return err
		}
	}

	// turn results into xml (bytes form)
	defer cursor.Close(*ctx)
	xmlEncoding := []byte(xml.Header)
	xmlEncoding = append(xmlEncoding, []byte("<Log>\n")...)
	for cursor.Next(*ctx) {
		event := &event.Event{}
		err := cursor.Decode(event)
		if err != nil { 
			log.Printf("Error while decoding a mongo doc into go struct. : %v ", cursor.Current)
			continue 
			// not sure if this continue statement will work
		}

		eventBytes, err := xml.MarshalIndent(event.Data, "  ", "  ")
		if err != nil {
			log.Printf("Couldn't marshal parsed event %+v in to xml, error: %s", event.Data, err)
			continue
			// not sure if this continue statement will work
		}

		xmlEncoding = append(xmlEncoding, eventBytes...)
		xmlEncoding = append(xmlEncoding, []byte("\n")...)	
	}
	if err := cursor.Err(); err != nil {
		panic(err)
	}

	xmlEncoding = append(xmlEncoding, []byte("</Log>\n")...)
	log.Println("\n", string(xmlEncoding))


	// push bytes to rabbitMQ

	return nil
}

func handle(ctx *context.Context, command *Command) {
	err := handlerMap[command.Command](ctx, command)
	if err != nil {
		log.Printf("Error handling command %+v, error: %s", command, err)
	}

	// we should retry if handling fails.
}
