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
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func CreateUserAccount(ctx *context.Context, username string) (*UserAccount, error) {
	account := &UserAccount{
		Username:     username,
		Balance:      0,
		Created:      time.Now().Unix(),
		Updated:      time.Now().Unix(),
		BuyAmounts:   map[string]float64{},
		SellAmounts:  map[string]float64{},
		BuyTriggers:  []*Trigger{},
		SellTriggers: []*Trigger{},
		Stocks:       map[string]float64{},
		Transactions: []*Transaction{},
		RecentBuy:    &CommandHistory{},
		RecentSell:   &CommandHistory{},
	}

	bsonBytes, err := bson.Marshal(account)
	if err != nil {
		return nil, err
	}

	accountsCollection := client.Database("test").Collection("Accounts")
	_, err = accountsCollection.InsertOne(*ctx, bsonBytes)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func add(ctx *context.Context, command *Command) ([]byte, error) {
	account, err := find_account(ctx, command.Username)
	if err == mongo.ErrNoDocuments {
		account, err = CreateUserAccount(ctx, command.Username)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("failed to add funds for %s, error: %s", command.Username, err.Error())
	}
	account.Balance += command.Amount

	update := bson.M{"$set": bson.D{primitive.E{Key: "balance", Value: account.Balance}}}

	err = updateUserAccount(ctx, account.Username, update)
	if err != nil {
		return []byte{}, err
	}

	return []byte("successfully added funds to user account"), nil
}

func commit_buy(ctx *context.Context, command *Command) ([]byte, error) {
	account, err := find_account(ctx, command.Username)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to commit buy for %s, error: %s", command.Username, err.Error())
	}

	stock := account.RecentBuy.stock
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
	account.RecentSell.stock = ""

	update := bson.M{"$set": bson.D{primitive.E{Key: "recentSell", Value: account.RecentSell}}}

	err = updateUserAccount(ctx, account.Username, update)
	if err != nil {
		return []byte{}, err
	}

	return []byte("Successfully cancelled the recent BUY"), nil
}

func commit_sell(ctx *context.Context, command *Command) ([]byte, error) {
	return []byte{}, nil
}

func cancel_sell(ctx *context.Context, command *Command) ([]byte, error) {
	account, err := find_account(ctx, command.Username)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to cancel buy for %s, error: %s", command.Username, err.Error())
	}

	account.RecentSell.Timestamp = 0
	account.RecentSell.Amount = 0
	account.RecentSell.stock = ""

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

	account.RecentBuy.Amount = command.Amount
	account.RecentBuy.Timestamp = time.Now().Unix()
	account.RecentBuy.stock = command.Stock

	update := bson.M{"$set": bson.D{primitive.E{Key: "recentBuy", Value: account.RecentBuy}}}

	err = updateUserAccount(ctx, account.Username, update)
	if err != nil {
		return []byte{}, err
	}

	return []byte("buy command successful"), nil

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
			return nil, err
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

func display_summary(ctx *context.Context, command *Command) ([]byte, error) {
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

	usernameEmpty := command.Username == ""
	stockSymbolEmpty := command.Stock == ""

	parseErrors.usernameEmpty = usernameEmpty
	parseErrors.stockSymbolEmpty = stockSymbolEmpty

	parseErrors.AmountNotConvertibleToFloat = true
	return nil
}
