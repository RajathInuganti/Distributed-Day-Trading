package main

import (
	"context"
	"log"
	"time"
)

func logUserCommandEvent(ctx *context.Context, transactionNum int64, server string, command *Command) {
	data := &UserCommand{
		Timestamp:      time.Now().Unix(),
		Server:         server,
		TransactionNum: transactionNum,
		Command:        command.Command,
		Username:       command.Username,
		StockSymbol:    command.Stock,
		Filename:       command.Filename,
		Funds:          command.Amount,
	}
	event := &Event{EventType: EventUserCommand, Data: data}
	insertEventToDB(ctx, event)
}

func logQuoteServerEvent(ctx *context.Context, transactionNum int64,
	server string, cryptokey string, quoteServerTime int64, command *Command) {
	data := &QuoteServer{
		Timestamp:       time.Now().Unix(),
		Server:          server,
		TransactionNum:  transactionNum,
		StockSymbol:     command.Stock,
		Username:        command.Username,
		QuoteServerTime: quoteServerTime,
		Cryptokey:       cryptokey,
	}
	event := &Event{EventType: EventQuoteServer, Data: data}
	insertEventToDB(ctx, event)
}

func logAccountTransactionEvent(ctx *context.Context, transactionNum int64, server, action string, command *Command) {
	data := &AccountTransaction{
		Timestamp:      time.Now().Unix(),
		Server:         server,
		TransactionNum: transactionNum,
		Action:         action,
		Username:       command.Username,
		Funds:          command.Amount,
	}
	event := &Event{EventType: EventAccountTransaction, Data: data}
	insertEventToDB(ctx, event)
}

func logSystemEvent(ctx *context.Context, transactionNum int64, server string, command *Command) {
	data := &SystemEvent{
		Timestamp:      time.Now().Unix(),
		Server:         server,
		TransactionNum: transactionNum,
		Command:        command.Command,
		Username:       command.Username,
		StockSymbol:    command.Stock,
		Filename:       command.Filename,
		Funds:          command.Amount,
	}
	event := &Event{EventType: EventSystem, Data: data}
	insertEventToDB(ctx, event)
}

func logErrorEvent(ctx *context.Context,transactionNum int64, server string, errorMsg string, command *Command) {
	data := &ErrorEvent{
		Timestamp:      time.Now().Unix(),
		Server:         server,
		TransactionNum: transactionNum,
		Command:        command.Command,
		Username:       command.Username,
		StockSymbol:    command.Stock,
		Filename:       command.Filename,
		ErrorMessage:   errorMsg,
		Funds:          command.Amount,
	}
	event := &Event{EventType: EventError, Data: data}
	insertEventToDB(ctx, event)
}

func logDebugEvent(ctx *context.Context, transactionNum int64, server, debugMsg string, command *Command) {
	data := &DebugEvent{
		Timestamp:      time.Now().Unix(),
		Server:         server,
		TransactionNum: transactionNum,
		Command:        command.Command,
		Username:       command.Username,
		StockSymbol:    command.Stock,
		Filename:       command.Filename,
		DebugMessage:   debugMsg,
		Funds:          command.Amount,
	}

	event := &Event{EventType: EventDebug, Data: data}
	insertEventToDB(ctx, event)
}

func insertEventToDB(ctx *context.Context, event *Event) {
	eventsCollection := client.Database("test").Collection("events")

	maxAttempts := 5
	for i := 1; i <= maxAttempts; i++ {
		_, err := eventsCollection.InsertOne(*ctx, event)
		if err == nil {
			return
		}

		log.Printf("Error inserting event to DB: %v, err: %s, attempt: %d", event, err, i)
		if i == maxAttempts {
			log.Printf("Failed to insert data to DB after %d attempts", maxAttempts)
			return
		}
		time.Sleep(time.Minute)
	}
}
