package main

import (
	"encoding/xml"
	"log"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

const (
	EventUserCommand        = "userCommand"
	EventQuoteServer        = "quoteServer"
	EventAccountTransaction = "accountTransaction"
	EventSystem             = "systemEvent"
	EventError              = "errorEvent"
	EventDebug              = "debugEvent"
)

type requestData struct {
	Command  string `json:"Command"`
	Username string `json:"Username"`
	Amount   string `json:"Amount"`
	Stock    string `json:"Stock"`
	Filename string `json:"Filename"`
}

type Command struct {
	Command           string  `json:"Command"`
	Username          string  `json:"Username"`
	Amount            float64 `json:"Amount"`
	Stock             string  `json:"Stock"`
	Filename          string  `json:"Filename"`
	TransactionNumber int64   `json:"transactionNumber"`
}

func fromRequestDataToCommand(r *requestData) *Command {
	val := strings.TrimSuffix(r.Amount, "\r")

	amount, err := strconv.ParseFloat(val, 64)
	if err != nil {
		amount = 0
	}
	return &Command{
		Command:  r.Command,
		Username: r.Username,
		Amount:   amount,
		Stock:    r.Stock,
		Filename: r.Filename,
	}
}

type Response struct {
	Command string `json:"command"`
	Data    []byte `json:"data"`
	Error   string `json:"error"`
}

type ParsingErrors struct {
	usernameEmpty               bool
	stockSymbolEmpty            bool
	AmountNotConvertibleToFloat bool
}

type Transaction struct {
	ID              int64   `bson:"id"`
	Timestamp       int64   `bson:"timestamp"`
	TransactionType string  `bson:"transactionType"`
	Amount          float64 `bson:"amount"`
	Stock           string  `bson:"stock"`
}

type UserAccount struct {
	Username     string             `bson:"username"`
	Balance      float64            `bson:"balance"`
	Created      int64              `bson:"created"`
	Updated      int64              `bson:"updated"`
	BuyAmounts   map[string]float64 `bson:"buyAmounts"`
	SellAmounts  map[string]float64 `bson:"sellAmounts"`
	BuyTriggers  map[string]float64 `bson:"buyTriggers"`
	SellTriggers map[string]float64 `bson:"sellTriggers"`
	Stocks       map[string]float64 `bson:"stocks"`
	Transactions []*Transaction     `bson:"transactions"`
	RecentBuy    *CommandHistory    `bson:"recentBuy"`
	RecentSell   *CommandHistory    `bson:"recentSell"`
}

type CommandHistory struct {
	Timestamp int64   `bson:"timestamp"`
	Amount    float64 `bson:"amount"`
	Stock     string  `bson:"stock"`
}

// Event struct describes any 'event' that occurs in the system (any of UserCommand, QuoteServer, AccountTransaction, SystemEvent, ErrorEvent)
type Event struct {
	EventType string      `bson:"eventType"`
	Data      interface{} `bson:"data"`
}

// UnmarshalBSONValue is an implementation that helps in decoding MongoDB bson response to golang struct
func (e *Event) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	var rawData bson.Raw

	err := bson.Unmarshal(data, &rawData)
	if err != nil {
		log.Printf("Error unmarshalling bytes to type bson.RAW: %s, error: %s", string(data), err)
		return err
	}

	err = rawData.Lookup("eventType").Unmarshal(&e.EventType)
	if err != nil {
		log.Printf("Error unmarshalling eventType from rawBson: %+v, error: %s", rawData, err)
	}

	switch e.EventType {
	case EventUserCommand:
		e.EventType = EventUserCommand
		e.Data = &UserCommand{}
	case EventQuoteServer:
		e.EventType = EventQuoteServer
		e.Data = &QuoteServer{}
	case EventAccountTransaction:
		e.EventType = EventAccountTransaction
		e.Data = &AccountTransaction{}
	case EventSystem:
		e.EventType = EventSystem
		e.Data = &SystemEvent{}
	case EventError:
		e.EventType = EventError
		e.Data = &ErrorEvent{}
	case EventDebug:
		e.EventType = EventDebug
		e.Data = &DebugEvent{}
	default:
		log.Printf("Unknown eventType: %s", e.EventType)
	}

	err = rawData.Lookup("data").Unmarshal(e.Data)
	if err != nil {
		log.Printf("Couldn't Marshal rawBson : %+v, got error: %s", rawData, err)
		return err
	}

	return nil
}

// UserCommand: Any command issued by the user
type UserCommand struct {
	XMLName        xml.Name `xml:"userCommand"`
	Timestamp      int64    `xml:"timestamp"`
	Server         string   `xml:"server"`
	TransactionNum int64    `xml:"transactionNum"`
	Command        string   `xml:"command"`
	Username       string   `xml:"username"`
	StockSymbol    string   `xml:"stockSymbol"`
	Filename       string   `xml:"filename"`
	Funds          float64  `xml:"funds"`
}

// QuoteServer: Any communication with the quoter server
type QuoteServer struct {
	XMLName         xml.Name `xml:"quoteServer"`
	Timestamp       int64    `xml:"timestamp"`
	Server          string   `xml:"server"`
	TransactionNum  int64    `xml:"transactionNum"`
	Price           float64  `xml:"price"`
	StockSymbol     string   `xml:"stockSymbol"`
	Username        string   `xml:"username"`
	QuoteServerTime int64    `xml:"quoteServerTime"`
	Cryptokey       string   `xml:"cryptokey"`
}

// AccountTransaction: any change in User's account
type AccountTransaction struct {
	XMLName        xml.Name `xml:"accountTransaction"`
	Timestamp      int64    `xml:"timestamp"`
	Server         string   `xml:"server"`
	TransactionNum int64    `xml:"transactionNum"`
	Action         string   `xml:"action"`
	Username       string   `xml:"username"`
	Funds          float64  `xml:"funds"`
}

// SystemEvent: Any event that is triggered by our system. For example, buying a stock because a trigger was set by the user.
type SystemEvent struct {
	XMLName        xml.Name `xml:"systemEvent"`
	Timestamp      int64    `xml:"timestamp"`
	Server         string   `xml:"server"`
	TransactionNum int64    `xml:"transactionNum"`
	Command        string   `xml:"command"`
	Username       string   `xml:"username"`
	StockSymbol    string   `xml:"stockSymbol"`
	Filename       string   `xml:"filename"`
	Funds          float64  `xml:"funds"`
}

// ErrorEvent: Any error that occurs for a transaction with the quote server
type ErrorEvent struct {
	XMLName        xml.Name `xml:"errorEvent"`
	Timestamp      int64    `xml:"timestamp"`
	Server         string   `xml:"server"`
	TransactionNum int64    `xml:"transactionNum"`
	Command        string   `xml:"command"`
	Username       string   `xml:"username"`
	StockSymbol    string   `xml:"stockSymbol"`
	Filename       string   `xml:"filename"`
	Funds          float64  `xml:"funds"`
	ErrorMessage   string   `xml:"errorMessage"`
}

// Debug: debug logs for ourselves
type DebugEvent struct {
	XMLName        xml.Name `xml:"debug"`
	Timestamp      int64    `xml:"timestamp"`
	Server         string   `xml:"server"`
	TransactionNum int64    `xml:"transactionNum"`
	Command        string   `xml:"command"`
	Username       string   `xml:"username"`
	StockSymbol    string   `xml:"stockSymbol"`
	Filename       string   `xml:"filename"`
	Funds          float64  `xml:"funds"`
	DebugMessage   string   `xml:"debugMessage"`
}
