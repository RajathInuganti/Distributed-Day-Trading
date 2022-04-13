package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	//"os"
	//"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var parseErrors ParsingErrors
var transactionNumber int64
var txMutex sync.Mutex

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

func getTransactionNumber() int64 {
	txMutex.Lock()
	defer txMutex.Unlock()

	transactionNumber += 1
	return transactionNumber
}

func add(ctx *context.Context, command *Command) ([]byte, error) {
	account, err := find_account(ctx, command.Username)
	if err == mongo.ErrNoDocuments {
		account, err = CreateUserAccount(ctx, command.Username)
		if err != nil {
			return nil, err
		}
	}

	if err != nil && err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("failed to add funds for %s, error: %s", command.Username, err.Error())
	}

	account.Balance += command.Amount

	update := bson.M{"$set": bson.M{"balance": account.Balance}}

	err = updateUserAccount(ctx, account.Username, update)
	if err != nil {
		return []byte{}, err
	}

	go logAccountTransactionEvent(ctx, getHostname(), "add", command)
	return []byte("successfully added funds to user account"), nil
}

func commit_buy(ctx *context.Context, command *Command) ([]byte, error) {
	account, err := find_account(ctx, command.Username)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to commit buy for %s, error: %s", command.Username, err.Error())
	}

	stock := account.RecentBuy.Stock
	stock_amount := account.RecentBuy.Amount
	buy_timestamp := account.RecentBuy.Timestamp

	time_elapsed := time.Now().Unix() - buy_timestamp

	if time_elapsed <= 60 {
		account.Balance -= stock_amount
		account.Stocks[stock] += stock_amount

		account.RecentBuy.Amount = 0
		account.RecentBuy.Stock = ""
		account.RecentBuy.Timestamp = 0

		update := bson.M{
			"$set": bson.M{
				"balance":   account.Balance,
				"stocks":    account.Stocks,
				"recentBuy": account.RecentBuy,
			},
		}

		err = updateUserAccount(ctx, account.Username, update)
		if err != nil {
			return []byte{}, err
		}

		go logAccountTransactionEvent(ctx, getHostname(), "remove", command)
		return []byte("successfully committed the most recent buy"), nil

	}

	return nil, errors.New("commit buy executed after 60 seconds, or no buy was commited - failed")
}

func cancel_buy(ctx *context.Context, command *Command) ([]byte, error) {
	account, err := find_account(ctx, command.Username)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to cancel buy for %s, error: %s", command.Username, err.Error())
	}

	account.RecentBuy.Timestamp = 0
	account.RecentBuy.Amount = 0
	account.RecentSell.Stock = ""

	update := bson.M{"$set": bson.D{primitive.E{Key: "recentSell", Value: account.RecentSell}}}

	err = updateUserAccount(ctx, account.Username, update)
	if err != nil {
		return []byte{}, err
	}

	return []byte("Successfully cancelled the recent BUY"), nil
}

func commit_sell(ctx *context.Context, command *Command) ([]byte, error) {
	account, err := find_account(ctx, command.Username)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to commit buy for %s, error: %s", command.Username, err.Error())
	}

	stock := account.RecentSell.Stock
	stock_amount := account.RecentSell.Amount
	sell_timestamp := account.RecentSell.Timestamp

	time_elapsed := time.Now().Unix() - sell_timestamp

	if time_elapsed <= 60 {
		account.Balance += stock_amount
		account.Stocks[stock] -= stock_amount

		if account.Stocks[stock] == 0 {
			delete(account.Stocks, stock)
		}

		account.RecentSell.Amount = 0
		account.RecentSell.Stock = ""
		account.RecentSell.Timestamp = 0

		update := bson.M{
			"$set": bson.M{
				"balance":    account.Balance,
				"stocks":     account.Stocks,
				"recentSell": account.RecentSell,
			},
		}

		err = updateUserAccount(ctx, account.Username, update)
		if err != nil {
			return []byte{}, err
		}

		go logAccountTransactionEvent(ctx, getHostname(), "add", command)

		return []byte("successfully committed the most recent sell"), nil

	}
	return nil, errors.New("commit sell executed after 60 seconds - failed or prior sell not executed")
}

func cancel_sell(ctx *context.Context, command *Command) ([]byte, error) {
	account, err := find_account(ctx, command.Username)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to cancel buy for %s, error: %s", command.Username, err.Error())
	}

	account.RecentSell.Timestamp = 0
	account.RecentSell.Amount = 0
	account.RecentSell.Stock = ""

	update := bson.M{"$set": bson.D{primitive.E{Key: "recentSell", Value: account.RecentSell}}}

	err = updateUserAccount(ctx, account.Username, update)
	if err != nil {
		return []byte{}, err
	}

	return []byte("Successfully cancelled the recent SELL"), nil
}

func buy(ctx *context.Context, command *Command) ([]byte, error) {
	account, err := find_account(ctx, command.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to buy selected stock for %s, error: %s", command.Username, err.Error())
	}

	if account.Balance < command.Amount {
		return nil, errors.New("buy failed - insufficient funds")
	}

	account.RecentBuy.Amount = command.Amount
	account.RecentBuy.Timestamp = time.Now().Unix()
	account.RecentBuy.Stock = command.Stock

	update := bson.M{
		"$set": bson.M{
			"recentBuy": account.RecentBuy,
		}}

	err = updateUserAccount(ctx, account.Username, update)
	if err != nil {
		return []byte{}, err
	}

	return []byte("buy command successful"), nil

}

func sell(ctx *context.Context, command *Command) ([]byte, error) {
	account, err := find_account(ctx, command.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to sell selected stock for %s, error: %s", command.Username, err.Error())
	}

	if account.Stocks[command.Stock] >= command.Amount {
		account.RecentSell.Amount = command.Amount
		account.RecentSell.Timestamp = time.Now().Unix()
		account.RecentSell.Stock = command.Stock

		update := bson.M{"$set": bson.D{primitive.E{Key: "recentSell", Value: account.RecentSell}}}

		err = updateUserAccount(ctx, account.Username, update)
		if err != nil {
			return []byte{}, err
		}

		return []byte("sell command successful"), nil
	}
	return nil, errors.New("sell failed - insufficient amount of selected stock")
}

func set_buy_amount(ctx *context.Context, command *Command) ([]byte, error) {

	account, err := find_account(ctx, command.Username)
	if err != nil {
		return nil, err
	}

	_, found := account.BuyAmounts[command.Stock]
	if found {
		if account.BuyAmounts[command.Stock] > 0 {
			account.Balance = account.Balance + account.BuyAmounts[command.Stock]
			account.BuyAmounts[command.Stock] = 0
		}
	}

	if account.Balance >= command.Amount {
		account.Balance = account.Balance - command.Amount
		account.BuyAmounts[command.Stock] = command.Amount

		update := bson.M{"$set": bson.M{
			"balance":    account.Balance,
			"buyAmounts": account.BuyAmounts,
		},
		}

		err = updateUserAccount(ctx, command.Username, update)
		if err != nil {
			return []byte{}, err
		}

		return []byte("successfully set aside buy amount"), nil
	}

	return nil, errors.New("not enough account balance")
}

func set_buy_trigger(ctx *context.Context, command *Command) ([]byte, error) {

	var price_adjustment bool = false

	account, err := find_account(ctx, command.Username)
	if err != nil {
		return nil, err
	}

	if account.BuyAmounts[command.Stock] >= command.Amount {

		price, found := account.BuyTriggers[command.Stock]
		if found {
			price_adjustment = true
		}

		account.BuyTriggers[command.Stock] = command.Amount
		update := bson.M{
			"$set": bson.M{
				"buyTriggers": account.BuyTriggers,
			},
		}
		err := updateUserAccount(ctx, command.Username, update)
		if err != nil {
			log.Printf("Error updating account")
		}

		return trigger(ctx, command, price_adjustment, price, "BUY"), nil

	}

	return nil, errors.New("not enough buy amount for transaction")
}

func set_sell_amount(ctx *context.Context, command *Command) ([]byte, error) {

	account, err := find_account(ctx, command.Username)
	if err != nil {
		return nil, err
	}

	_, found := account.SellAmounts[command.Stock]
	if found {
		if account.SellAmounts[command.Stock] > 0 {
			account.Stocks[command.Stock] = account.Stocks[command.Stock] + account.SellAmounts[command.Stock]
			account.SellAmounts[command.Stock] = 0
		}
	}

	if account.Stocks[command.Stock] >= command.Amount {
		account.Stocks[command.Stock] = account.Stocks[command.Stock] - command.Amount
		account.SellAmounts[command.Stock] = command.Amount

		update := bson.M{"$set": bson.M{
			"stocks":      account.Stocks,
			"sellAmounts": account.SellAmounts,
		},
		}

		err = updateUserAccount(ctx, command.Username, update)
		if err != nil {
			return []byte{}, err
		}

		return []byte("successfully set aside sell amount"), nil
	}

	return nil, errors.New("not enough stock balance")
}

func set_sell_trigger(ctx *context.Context, command *Command) ([]byte, error) {

	var price_adjustment bool = false

	account, err := find_account(ctx, command.Username)
	if err != nil {
		return nil, err
	}

	if account.SellAmounts[command.Stock] >= command.Amount {

		price, found := account.SellTriggers[command.Stock]
		if found {
			price_adjustment = true
		}

		account.SellTriggers[command.Stock] = command.Amount
		update := bson.M{
			"$set": bson.M{
				"sellTriggers": account.SellTriggers,
			},
		}

		err := updateUserAccount(ctx, command.Username, update)
		if err != nil {
			log.Printf("Error updating account")
		}

		return trigger(ctx, command, price_adjustment, price, "SELL"), nil

	}

	return nil, errors.New("not enough sell amount stock for transaction")
}

func quote(ctx *context.Context, command *Command) ([]byte, error) {
	if command.Stock == "" || command.Username == "" {
		return nil, errors.New("quote command requires stock and username")
	}

	result, err := get_quote(command.Stock, command.Username)
	if err != nil {
		return nil, err
	}

	price, timestamp, cryptoKey, err := parseQuote(result)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote for %s, error: %s", command.Username, err.Error())
	}

	go logQuoteServerEvent(ctx, getHostname(), cryptoKey, timestamp, price, command)

	responseString := fmt.Sprintf("\nstock %s\n: price %.2f\n\n", command.Stock, price)
	log.Printf("Quote handler response: %s\n", responseString)
	return []byte(responseString), nil
}

func cancel_set_buy(ctx *context.Context, command *Command) ([]byte, error) {
	if command.Username == "" {
		return nil, errors.New("username is required for cancel_set_buy")
	}

	account, err := find_account(ctx, command.Username)
	if err != nil {
		return nil, err
	}

	_, found := account.BuyAmounts[command.Stock]
	if !found {
		return nil, errors.New("no previous buy amount set")
	}

	account.Balance += account.BuyAmounts[command.Stock]
	command.Amount = account.BuyAmounts[command.Stock]
	delete(account.BuyAmounts, command.Stock)

	update := bson.M{
		"$set": bson.M{
			"balance":    account.Balance,
			"buyAmounts": account.BuyAmounts,
		},
	}

	err = updateUserAccount(ctx, account.Username, update)
	if err != nil {
		return []byte{}, err
	}

	go logAccountTransactionEvent(ctx, getHostname(), "add", command)
	return []byte("Successfully cancelled the SET_BUY_AMOUNT"), nil
}

func cancel_set_sell(ctx *context.Context, command *Command) ([]byte, error) {
	if command.Username == "" {
		return nil, errors.New("username is required for cancel_set_sell")
	}

	account, err := find_account(ctx, command.Username)
	if err != nil {
		return nil, err
	}

	_, found := account.SellAmounts[command.Stock]
	if !found {
		return nil, errors.New("no previous sell amount set")
	}

	account.Stocks[command.Stock] += account.SellAmounts[command.Stock]
	command.Amount = account.SellAmounts[command.Stock]
	delete(account.SellAmounts, command.Stock)

	update := bson.M{
		"$set": bson.M{
			"sellAmounts": account.SellAmounts,
			"stocks":      account.Stocks,
		},
	}

	err = updateUserAccount(ctx, account.Username, update)
	if err != nil {
		return []byte{}, err
	}

	go logAccountTransactionEvent(ctx, getHostname(), "add", command)
	return []byte("Successfully cancelled the SET_SELL_AMOUNT"), nil
}

func display_summary(ctx *context.Context, command *Command) ([]byte, error) {

	if command.Username == "" {
		return nil, errors.New("username is required for DISPLAY_SUMMARY")
	}

	account, err := find_account(ctx, command.Username)
	if err != nil {
		return nil, err
	}

	summary := "-----User Account Summary-----\n"
	summary += fmt.Sprintf("Username: %s\n", account.Username)
	summary += fmt.Sprintf("balance: %f\n", account.Balance)
	for stock, amount := range account.Stocks {
		summary += fmt.Sprintf("stock %s: %f\n", stock, amount)
	}
	for _, t := range account.Transactions {
		summary += fmt.Sprintf("transaction: %3d, %9d, %s, %s, %f\n", t.ID, t.Timestamp, t.TransactionType, t.Stock, t.Amount)
	}
	for _, t := range account.BuyTriggers {
		summary += fmt.Sprintf("buy trigger: %v\n", t)
	}
	for _, t := range account.SellTriggers {
		summary += fmt.Sprintf("sell trigger: %v\n", t)
	}
	summary += "-----End------\n\n"

	return []byte(summary), nil
}

func dumplog(ctx *context.Context, command *Command) ([]byte, error) {

	eventCollection := client.Database("test").Collection("events")

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
	xmlEncoding = append(xmlEncoding, []byte("<log>\n")...)
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

	xmlEncoding = append(xmlEncoding, []byte("</log>\n")...)

	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	_, err = gz.Write(xmlEncoding)
	if err != nil {
		log.Printf("Error while writing to gzip writer, error: %s", err)
		return []byte{}, err
	}

	gz.Close()

	return b.Bytes(), nil
}

func handle(ctx *context.Context, data []byte) *Response {
	requestDataStruct := &requestData{}

	err := json.Unmarshal(data, requestDataStruct)
	if err != nil {
		log.Printf("Failed to unmarshal message: %s, error: %s", string(data), err.Error())
		return &Response{Data: []byte{}, Error: "Invalid data sent"}
	}

	command := fromRequestDataToCommand(requestDataStruct)

	log.Printf("Received command: %+v", command)
	response := &Response{}
	command.TransactionNumber = getTransactionNumber()
	err = verifyAndParseRequestData(command)
	if err != nil {
		response.Error = err.Error()
		go logErrorEvent(ctx, getHostname(), err.Error(), command)
		return response
	}

	go logUserCommandEvent(ctx, getHostname(), command)

	responseData, err := handlerMap[command.Command](ctx, command)
	if err != nil {
		log.Printf("Error handling command %+v, error: %s", command, err)
		response.Error = err.Error()
		go logErrorEvent(ctx, getHostname(), err.Error(), command)
		return response
	}

	response.Data = responseData
	response.Command = command.Command
	return response
}

func verifyAndParseRequestData(command *Command) error {
	if command.Command == "" {
		return errors.New("no command specified")
	}

	usernameEmpty := command.Username == ""
	stockSymbolEmpty := command.Stock == ""

	parseErrors.usernameEmpty = usernameEmpty
	parseErrors.stockSymbolEmpty = stockSymbolEmpty

	parseErrors.AmountNotConvertibleToFloat = true
	return nil
}

func getHostname() string {
	hostname, errHostname := os.Hostname()
	if errHostname != nil {
		hostname = "txServerMain"
	}

	return hostname
}
