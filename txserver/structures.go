package main

import (
	"log"

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

type Command struct {
	Command  string  `json:"Command"`
	Username string  `json:"Username"`
	Amount   float32 `json:"Amount"`
	Stock    string  `json:"Stock"`
	Filename string  `json:"Filename"`
}

type Response struct {
	Data  []byte `json:"data"`
	Error string `json:"error"`
}

type ParsingErrors struct {
	usernameEmpty               bool
	stockSymbolEmpty            bool
	AmountNotConvertibleToFloat bool
}

type Transaction struct {
	ID              int64   `json:"transactionNum"`
	Timestamp       int64   `json:"timestamp"`
	TransactionType string  `json:"transactionType"`
	Amount          float32 `json:"amount"`
	Stock           string  `json:"stock"`
}

type UserAccount struct {
	username       string                        `bson:"username"`
	balance        float32                       `bson:"balance"`
	created        int                           `bson:"created"`
	updated        int                           `bson:"updated"`
	setBuyAmounts  map[string]float32            `bson:"buy"`
	setSellAmounts map[string]float32            `bson:"sell"`
	buyTriggers    map[string]map[float32]string `bson:"buyTriggers"`
	sellTriggers   map[string]map[float32]string `bson:"sellTriggers"`
	stocks         map[string]float32            `bson:"stocks"`
	transactions   []*Transaction                `bson:"transactions"`
	recentBuy      *CommandHistory               `bson:"recentBuy"`
	recentSell     *CommandHistory               `bson:"recentSell"`
}

type CommandHistory struct {
	Timestamp int64   `bson:"timestamp"`
	Amount    float32 `bson:"amount"`
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
		e.Data = UserCommand{}
	case EventQuoteServer:
		e.EventType = EventQuoteServer
		e.Data = QuoteServer{}
	case EventAccountTransaction:
		e.EventType = EventAccountTransaction
		e.Data = AccountTransaction{}
	case EventSystem:
		e.EventType = EventSystem
		e.Data = SystemEvent{}
	case EventError:
		e.EventType = EventError
		e.Data = ErrorEvent{}
	case EventDebug:
		e.EventType = EventDebug
		e.Data = DebugEvent{}
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
	Timestamp      int64   `xml:"timestamp"`
	Server         string  `xml:"server"`
	TransactionNum int64   `xml:"transactionNum"`
	Command        string  `xml:"command"`
	Username       string  `xml:"username"`
	StockSymbol    string  `xml:"stockSymbol"`
	Filename       string  `xml:"filename"`
	Funds          float32 `xml:"funds"`
}

// QuoteServer: Any communication with the quoter server
type QuoteServer struct {
	Timestamp       int64   `xml:"timestamp"`
	Server          string  `xml:"server"`
	TransactionNum  int64   `xml:"transactionNum"`
	Price           float64 `xml:"price"`
	StockSymbol     string  `xml:"stockSymbol"`
	Username        string  `xml:"username"`
	QuoteServerTime int64   `xml:"quoteServerTime"`
	Cryptokey       string  `xml:"cryptokey"`
}

// AccountTransaction: any change in User's account
type AccountTransaction struct {
	Timestamp      int64   `xml:"timestamp"`
	Server         string  `xml:"server"`
	TransactionNum int64   `xml:"transactionNum"`
	Action         string  `xml:"action"`
	Username       string  `xml:"username"`
	Funds          float32 `xml:"funds"`
}

// SystemEvent: Any event that is triggered by our system. For example, buying a stock because a trigger was set by the user.
type SystemEvent struct {
	Timestamp      int64   `xml:"timestamp"`
	Server         string  `xml:"server"`
	TransactionNum int64   `xml:"transactionNum"`
	Command        string  `xml:"command"`
	Username       string  `xml:"username"`
	StockSymbol    string  `xml:"stockSymbol"`
	Filename       string  `xml:"filename"`
	Funds          float32 `xml:"funds"`
}

// ErrorEvent: Any error that occurs for a transaction with the quote server
type ErrorEvent struct {
	Timestamp      int64   `xml:"timestamp"`
	Server         string  `xml:"server"`
	TransactionNum int64   `xml:"transactionNum"`
	Command        string  `xml:"command"`
	Username       string  `xml:"username"`
	StockSymbol    string  `xml:"stockSymbol"`
	Filename       string  `xml:"filename"`
	Funds          float32 `xml:"funds"`
	ErrorMessage   string  `xml:"errorMessage"`
}

// Debug: debug logs for ourselves
type DebugEvent struct {
	Timestamp      int64   `xml:"timestamp"`
	Server         string  `xml:"server"`
	TransactionNum int64   `xml:"transactionNum"`
	Command        string  `xml:"command"`
	Username       string  `xml:"username"`
	StockSymbol    string  `xml:"stockSymbol"`
	Filename       string  `xml:"filename"`
	Funds          float32 `xml:"funds"`
	DebugMessage   string  `xml:"debugMessage"`
}
