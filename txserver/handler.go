package main

import (
	"context"
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
	account, err := find_account(ctx, command.Username)
	if err != nil {
		return nil, err
	}

	account.Balance += command.Amount

	Accounts := client.Database("test").Collection("Accounts")
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"username": command.Username}

	result, err := Accounts.UpdateOne(context.TODO(), filter, account, opts)

	if err != nil {
		return nil, errors.New("account update unsuccessful")
	}

	if result.MatchedCount == 1 {
		return []byte("account balance updated"), nil
	}

	return nil, errors.New("account update unsuccessful")
}

func commit_buy(ctx *context.Context, command *Command) ([]byte, error) {
	account, err := find_account(ctx, command.Username)
	if err != nil {
		return nil, err
	}

	stock := account.RecentBuy.stock
	stock_amount := account.RecentBuy.Amount
	buy_timestamp := account.RecentBuy.Timestamp

	time_elapsed := time.Now().Unix() - buy_timestamp

	if time_elapsed <= 60 {
		account.Balance -= stock_amount
		account.Stocks[stock] += stock_amount

		Accounts := client.Database("test").Collection("Accounts")
		opts := options.Update().SetUpsert(true)
		filter := bson.M{"username": command.Username}

		result, err := Accounts.UpdateOne(context.TODO(), filter, account, opts)

		if err != nil {
			return nil, errors.New("commit buy unsuccessful")
		}

		if result.MatchedCount == 1 {
			return []byte("commit buy successful"), nil
		}
	}
	return nil, errors.New("commit buy executed after 60 seconds - failed")
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
	account, err := find_account(ctx, command.Username)
	if err != nil {
		return nil, err
	}

	account.RecentBuy.Amount = command.Amount
	account.RecentBuy.Timestamp = time.Now().Unix()
	account.RecentBuy.stock = command.Stock

	Accounts := client.Database("test").Collection("Accounts")
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"username": command.Username}

	result, err := Accounts.UpdateOne(context.TODO(), filter, account, opts)

	if err != nil {
		return nil, errors.New("buy command unsuccessful")
	}

	if result.MatchedCount == 1 {
		return []byte("buy command successful"), nil
	}

	return nil, errors.New("buy command unsuccessful")
}

func sell(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func find_account(ctx *context.Context, username string) (*UserAccount, error) {
	var account UserAccount

	if parseErrors.stockSymbolEmpty || parseErrors.usernameEmpty || parseErrors.AmountNotConvertibleToFloat {
		return &account, errors.New("insufficient information")
	}

	AccountsCollection := client.Database("test").Collection("Accounts")
	err := AccountsCollection.FindOne(*ctx, bson.M{"username": username}).Decode(&account)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("No account found for %s", username)
		}

		log.Printf("Error finding account with username: %s, error: %s", username, err.Error())
		return nil, errors.New("an Error occured while finding the account")
	}

	return &account, nil
}

func set_buy_amount(ctx *context.Context, command *Command) ([]byte, error) {

	account, err := find_account(ctx, command.Username)
	if err != nil {
		return nil, err
	}

	if account.BuyAmounts[command.Stock] > 0 {
		account.Balance = account.Balance + account.BuyAmounts[command.Stock]
		account.BuyAmounts[command.Stock] = 0
	}

	if account.Balance >= command.Amount {
		account.Balance = account.Balance - command.Amount
		account.BuyAmounts[command.Stock] = command.Amount

		opts := options.Update().SetUpsert(true)
		filter := bson.M{"username": command.Username}

		accountsCollection := client.Database("test").Collection("Accounts")
		result, err := accountsCollection.UpdateOne(context.TODO(), filter, account, opts)
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

	account, err := find_account(ctx, command.Username)
	if err != nil {
		return nil, err
	}

	if account.SellAmounts[command.Stock] > 0 {
		account.Stocks[command.Stock] = account.Stocks[command.Stock] + account.SellAmounts[command.Stock]
		account.BuyAmounts[command.Stock] = 0
	}

	if account.Stocks[command.Stock] >= command.Amount {
		account.Stocks[command.Stock] = account.Stocks[command.Stock] - command.Amount
		account.SellAmounts[command.Stock] = command.Amount

		opts := options.Update().SetUpsert(true)
		filter := bson.M{"username": command.Username}

		accountsCollection := client.Database("test").Collection("Accounts")
		result, err := accountsCollection.UpdateOne(context.TODO(), filter, account, opts)
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
		event := &Event{}
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
		logErrorEvent(ctx, 1, "server1", err.Error(), command)
		return response
	}

	logUserCommandEvent(ctx, 1, "server1", command)

	responseData, err := handlerMap[command.Command](ctx, command)
	if err != nil {
		log.Printf("Error handling command %+v, error: %s", command, err)
		response.Error = err.Error()
		logErrorEvent(ctx, 1, "server1", err.Error(), command)
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
