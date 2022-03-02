package main

import (
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
	"go.mongodb.org/mongo-driver/mongo/options"
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

	update := bson.M{"$set": bson.D{primitive.E{Key: "balance", Value: account.Balance}}}

	err = updateUserAccount(ctx, account.Username, update)
	if err != nil {
		return []byte{}, err
	}

	logAccountTransactionEvent(ctx, getHostname(), command.Command, command)
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

		update := bson.M{
			"$set": bson.M{
				"balance": account.Balance,
				"stocks":  account.Stocks,
			},
		}

		err = updateUserAccount(ctx, account.Username, update)
		if err != nil {
			return []byte{}, err
		}

		logAccountTransactionEvent(ctx, getHostname(), command.Command, command)
		return []byte("successfully committed the most recent buy"), nil

	}
	return nil, errors.New("commit buy executed after 60 seconds - failed")
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

	logAccountTransactionEvent(ctx, getHostname(), command.Command, command)

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

		update := bson.M{
			"$set": bson.M{
				"balance": account.Balance,
				"stocks":  account.Stocks,
			},
		}

		err = updateUserAccount(ctx, account.Username, update)
		if err != nil {
			return []byte{}, err
		}

		logAccountTransactionEvent(ctx, getHostname(), command.Command, command)

		return []byte("successfully committed the most recent sell"), nil

	}
	return nil, errors.New("commit sell executed after 60 seconds - failed")
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

	logAccountTransactionEvent(ctx, getHostname(), command.Command, command)
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

	update := bson.M{"$set": bson.D{primitive.E{Key: "recentBuy", Value: account.RecentBuy}}}

	err = updateUserAccount(ctx, account.Username, update)
	if err != nil {
		return []byte{}, err
	}

	logAccountTransactionEvent(ctx, getHostname(), command.Command, command)
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

		logAccountTransactionEvent(ctx, getHostname(), command.Command, command)
		return []byte("sell command successful"), nil
	}
	return nil, errors.New("sell failed - insufficient amount of selected stock")
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

	var price_adjustment bool = false

	account, err := find_account(ctx, command.Username)
	if err != nil {
		return nil, err
	}

	price, found := account.BuyTriggers[command.Stock]
	if found {
		price_adjustment = true
	}
	account.BuyTriggers[command.Stock] = float64(command.Amount)

	if account.BuyAmounts[command.Stock] >= command.Amount {
		update := bson.M{
			"$set": bson.M{
				"buyTriggers": account.BuyTriggers,
			},
		}
		updateUserAccount(ctx, command.Username, update)

		return trigger(ctx, command, price_adjustment, price, "BUY"), nil

	}

	return nil, errors.New("not enough amount for transaction")
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

	var price_adjustment bool = false

	account, err := find_account(ctx, command.Username)
	if err != nil {
		return nil, err
	}

	price, found := account.SellTriggers[command.Stock]
	if found {
		price_adjustment = true
	}
	account.SellTriggers[command.Stock] = float64(command.Amount)

	// check if user has set multiple price triggers for same stock

	if account.Stocks[command.Stock] >= command.Amount {
		update := bson.M{
			"$set": bson.M{
				"sellTriggers": account.SellTriggers,
			},
		}

		updateUserAccount(ctx, command.Username, update)
		return trigger(ctx, command, price_adjustment, price, "SELL"), nil

	}

	return nil, errors.New("not enough stock for transaction")
}

func quote(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func cancel_set_buy(ctx *context.Context, command *Command) ([]byte, error) {
	if command.Username == "" {
		return nil, errors.New("username is required for cancel_set_buy")
	}

	account, err := find_account(ctx, command.Username)
	if err != nil {
		return nil, err
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

	logAccountTransactionEvent(ctx, getHostname(), "ADD", command)
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

	account.Balance += account.SellAmounts[command.Stock]
	command.Amount = account.SellAmounts[command.Stock]
	delete(account.SellAmounts, command.Stock)

	update := bson.M{
		"$set": bson.M{
			"balance":     account.Balance,
			"sellAmounts": account.SellAmounts,
		},
	}

	err = updateUserAccount(ctx, account.Username, update)
	if err != nil {
		return []byte{}, err
	}

	logAccountTransactionEvent(ctx, getHostname(), "ADD", command)
	return []byte("Successfully cancelled the SET_SELL_AMOUNT"), nil
}

func display_summary(ctx *context.Context, command *Command) ([]byte, error) {
	log.Printf("display_summary ran")
	if command.Username == "" {
		return nil, errors.New("username is required for DISPLAY_SUMMARY")
	}

	account, err := find_account(ctx, command.Username)
	if err != nil {
		return nil, err
	}

	responseData, err := bson.Marshal(account)
	if err != nil {
		return nil, errors.New("an error occured while marshalling account")
	}

	return responseData, nil
}

func dumplog(ctx *context.Context, command *Command) ([]byte, error) {
	log.Printf("dumplog ran")
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
		logErrorEvent(ctx, getHostname(), err.Error(), command)
		return response
	}

	logUserCommandEvent(ctx, getHostname(), command)

	responseData, err := handlerMap[command.Command](ctx, command)
	if err != nil {
		log.Printf("Error handling command %+v, error: %s", command, err)
		response.Error = err.Error()
		logErrorEvent(ctx, getHostname(), err.Error(), command)
		return response
	}

	response.Data = responseData
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
