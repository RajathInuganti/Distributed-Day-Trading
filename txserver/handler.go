package main

import (
	"context"
	"day-trading/txserver/event"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	//"os"
	//"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var parseErrors ParsingErrors

var handlerMap = map[string]func(*context.Context, *Command) ([]byte, error){
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

func display_summary(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func buy(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func sell(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func find_account(ctx *context.Context, command *Command) (UserAccount, *mongo.Collection, error) {
	var account UserAccount

	if parseErrors.stockSymbolEmpty || parseErrors.usernameEmpty || parseErrors.AmountNotConvertibleToFloat {
		return account, nil, errors.New("insufficient information")
	}

	Accounts := client.Database("test").Collection("Accounts")
	err := Accounts.FindOne(*ctx, bson.M{"username": command.Username}).Decode(&account)
	if err != nil {
		return account, nil, errors.New("account not found")
	}

	return account, Accounts, nil
}

func set_buy_amount(ctx *context.Context, command *Command) ([]byte, error) {

	account, Accounts, err := find_account(ctx, command)
	if err != nil {
		return nil, err
	}

	if account.buy[command.Stock] > 0 {
		account.balance = account.balance + account.buy[command.Stock]
		account.buy[command.Stock] = 0
	}

	if account.balance >= command.Amount {
		account.balance = account.balance - command.Amount
		account.buy[command.Stock] = command.Amount

		opts := options.Update().SetUpsert(true)
		filter := bson.M{"username": command.Username}

		result, err := Accounts.UpdateOne(context.TODO(), filter, account, opts)
		if err != nil {
			return nil, errors.New("account update unsuccessful")
		}

		if result.MatchedCount == 1 {
			return []byte("Buy amount allocated"), nil
		}
	}

	return nil, errors.New("not enough account balance")
}

func set_buy_trigger(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func set_sell_amount(ctx *context.Context, command *Command) ([]byte, error) {

	account, Accounts, err := find_account(ctx, command)
	if err != nil {
		return nil, err
	}

	if account.sell[command.Stock] > 0 {
		account.stocks[command.Stock] = account.stocks[command.Stock] + account.sell[command.Stock]
		account.sell[command.Stock] = 0
	}

	if account.stocks[command.Stock] >= command.Amount {
		account.stocks[command.Stock] = account.stocks[command.Stock] - command.Amount
		account.sell[command.Stock] = command.Amount

		opts := options.Update().SetUpsert(true)
		filter := bson.M{"username": command.Username}

		result, err := Accounts.UpdateOne(context.TODO(), filter, account, opts)
		if err != nil {
			return nil, errors.New("account update unsuccessful")
		}

		if result.MatchedCount == 1 {
			return []byte("Sell amount allocated"), nil
		}
	}

	return nil, errors.New("not enough stock balance")
}

func set_sell_trigger(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func quote(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func cancel_set_buy(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func cancel_set_sell(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func dumplog(ctx *context.Context, command *Command) ([]byte, error) {
	eventCollection := client.Database("test").Collection("Events")

	// fetch results from mongo
	var cursor *mongo.Cursor
	var err error
	if command.Username != "" {
		// get events for specified user
		filter := bson.M{"data.username": command.Username}
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

func handle(ctx *context.Context, requestData []byte) *Response {
	command := &Command{}
	err := json.Unmarshal(requestData, command)
		if err != nil {
			log.Printf("Failed to unmarshal message: %+v", requestData)
			return &Response{Data: []byte{}, Error: "Invalid data sent"}
		}

	response := &Response{}
	err = verifyAndParseRequestData(command)
	if err != nil {
		response.Error = err.Error()
		logErrorEvent(ctx, time.Now().Unix(), 1, "server1", command.Command, command.Username, command.Stock, command.Filename, err.Error(), command.Amount)
		return response
	}

	logUserCommandEvent(ctx, time.Now().Unix(), 1, "server1", command.Command, command.Username, command.Stock, command.Filename, command.Amount)

	responseData, err := handlerMap[command.Command](ctx, command)
	if err != nil {
		log.Printf("Error handling command %+v, error: %s", command, err)
		response.Error = err.Error()
		logErrorEvent(ctx, time.Now().Unix(), 1, "server1", command.Command, command.Username, command.Stock, command.Filename, err.Error(), command.Amount)
		return response
	}

	response.Data = responseData
	return response
}

func verifyAndParseRequestData(command *Command) error {
	if command.Command == "" {
		return errors.New("no command specified")
	}

	amountFloatValue, AmountNotConvertibleToFloatError := strconv.ParseFloat(fmt.Sprintf("%v", command.Amount), 32)
	usernameEmpty := command.Username == ""
	stockSymbolEmpty := command.Stock == ""

	parseErrors.usernameEmpty = usernameEmpty
	parseErrors.stockSymbolEmpty = stockSymbolEmpty

	// setting Amount to an integer value so that it can be used in the rest of the code
	if AmountNotConvertibleToFloatError == nil {
		command.Amount = float32(amountFloatValue)
		return nil
	}

	parseErrors.AmountNotConvertibleToFloat = true
	return nil
}
