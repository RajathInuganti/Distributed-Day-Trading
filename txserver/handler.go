package main

import (
	"context"
	"day-trading/txserver/event"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"strconv"

	//"os"
	//"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	CONN_HOST = "192.168.4.2"
	CONN_PORT = 4444
)

var handlerMap = map[string]func(*context.Context, *Command) ([]byte, error) {
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

func add(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func commit_buy(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func cancel_buy(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func commit_sell(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func cancel_sell(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func buy(ctx *context.Context, command *Command) ([]byte, error)  {
	return []byte{}, nil
}

func sell(ctx *context.Context, command *Command) ([]byte, error)  {
	return []byte{}, nil
}

func set_buy_amount(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func set_buy_trigger(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func set_sell_amount(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func set_sell_trigger(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func quote(ctx *context.Context, command *Command) ([]byte,error)  {
	return []byte{}, nil
}

func cancel_set_buy(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func cancel_set_sell(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func display_summary(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func dumplog(ctx *context.Context, command *Command) ([]byte,error) {
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
			return []byte{}, err
		}
	} else {
		// get all events
		cursor, err = eventCollection.Find(*ctx, bson.D{})
		if err != nil {
			log.Printf("Error getting all the events from the Events collection, error:  %s", err)
			return []byte{}, err
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

	return xmlEncoding, nil
}

func handle(ctx *context.Context, command *Command) *Response{
	response := &Response{}
	responseData, err := handlerMap[command.Command](ctx, command)
	if err != nil {
		log.Printf("Error handling command %+v, error: %s", command, err)
		response.Error = err.Error()
		return response
	}

	response.Data = responseData
	return response
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

