package main

import (
	"context"
	"day-trading/txserver/event"
	"log"
	"time"
)

func logUserCommandEvent(
	ctx *context.Context, timestamp, transactionNum int64,
	server, command, username, stock, filename string, funds float64) error {
	data := &event.UserCommand{
		Timestamp:      timestamp,
		Server:         server,
		TransactionNum: transactionNum,
		Command:        command,
		Username:       username,
		StockSymbol:    stock,
		Filename:       filename,
		Funds:          funds,
	}
	event := &event.Event{EventType: event.EventUserCommand, Data: data}
	return nil
}

func logQuoteServerEvent(ctx *context.Context, timestamp, transactionNum int64,
	server, stock, username, cryptokey string, quoteServerTime int64) error {
	data := &event.QuoteServer{
		Timestamp:       timestamp,
		Server:          server,
		TransactionNum:  transactionNum,
		StockSymbol:     stock,
		Username:        username,
		QuoteServerTime: timestamp,
		Cryptokey:       cryptokey,
	}
	event := &event.Event{EventType: event.EventQuoteServer, Data: data}
	return nil
}

func logAccountTransactionEvent(ctx *context.Context, timestamp, transactionNum int64,
	server, action, username string, funds float64) error {
	data := &event.AccountTransaction{
		Timestamp:      timestamp,
		Server:         server,
		TransactionNum: transactionNum,
		Action:         action,
		Username:       username,
		Funds:          funds,
	}
	event := &event.Event{EventType: event.EventAccountTransaction, Data: data}
	return nil
}

func logSystemEvent(ctx *context.Context, timestamp, transactionNum int64,
	server, command, username, stock, filename string, funds float64) error {
	data := &event.SystemEvent{
		Timestamp:      timestamp,
		Server:         server,
		TransactionNum: transactionNum,
		Command:        command,
		Username:       username,
		StockSymbol:    stock,
		Filename:       filename,
		Funds:          funds,
	}
	event := &event.Event{EventType: event.EventSystem, Data: data}
	return nil
}

func logErrorEvent(ctx *context.Context, timestamp, transactionNum int64,
	server, command, username, stock, filename, errorMsg string, funds float64) error {
	data := &event.ErrorEvent{
		Timestamp:      timestamp,
		Server:         server,
		TransactionNum: transactionNum,
		Command:        command,
		Username:       username,
		StockSymbol:    stock,
		Filename:       filename,
		ErrorMessage:   errorMsg,
		Funds:          funds,
	}
	event := &event.Event{EventType: event.EventError, Data: data}
	return nil
}

func logDebugEvent(ctx *context.Context, timestamp, transactionNum int64,
	server, command, username, stock, filename, debugMsg string, funds float64) error {
	data := &event.DebugEvent{
		Timestamp:      timestamp,
		Server:         server,
		TransactionNum: transactionNum,
		Command:        command,
		Username:       username,
		StockSymbol:    stock,
		Filename:       filename,
		DebugMessage:   debugMsg,
		Funds:          funds,
	}

	event := &event.Event{EventType: event.EventDebug, Data: data}
	return nil
}

func insertEventToDB(ctx *context.Context, event *event.Event) error {
	eventsCollection := mongoClient.Database("test").Collection("events")
	_, err := eventsCollection.InsertOne(*ctx, event)
	if err != nil {
		log.Printf("Error inserting event to DB: %v, err: %s", event, err)
		return err
	}
	return nil
}

func retry(ctx *context.Context, maxAttempts int, sleep time.Duration, f func(ctx *context.Context, event *event.Event) error, event *event.Event) error {
	for i:= 1; i <= maxAttempts ; i++ {
		err := f(ctx, event)
		if err == nil {
			return nil
		}

		if maxAttempts == i {
			log.Printf("couldn't insert event to DB after %d attempts", i)
			return err
		}
		
		log.Printf("retry() failed attempt %d: %s", i, err)
		time.Sleep(sleep)
		sleep *= 2
		continue
	}

	return nil
}