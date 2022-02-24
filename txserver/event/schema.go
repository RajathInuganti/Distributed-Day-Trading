package event

import (
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

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
		log.Printf("Error unmarshalling bytes to type bson.RAW: %s, error: %s", string(data),  err)
		return err
	}

	err = rawData.Lookup("eventType").Unmarshal(&e.EventType)
	if err != nil {
		log.Printf("Error unmarshalling eventType from rawBson: %+v, error: %s", rawData,  err)
	}

	switch e.EventType {
	case "userCommand":
		e.EventType = "userCommand"
		e.Data = UserCommand{}
	case "quoteServer":
		e.EventType = "quoteServer"
		e.Data = QuoteServer{}
	case "accountTransaction":
		e.EventType = "accountTransaction"
		e.Data = AccountTransaction{}
	case "systemEvent":
		e.EventType = "systemEvent"
		e.Data = SystemEvent{}
	case "errorEvent":
		e.EventType = "errorEvent"
		e.Data = ErrorEvent{}
	case "debugEvent":
		e.EventType = "debug"
		e.Data = Debug{}
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
	Timestamp      int     `xml:"timestamp"`
	Server         string  `xml:"server"`
	TransactionNum int     `xml:"transactionNum"`
	Command        string  `xml:"command"`
	Username       string  `xml:"username"`
	StockSymbol    string  `xml:"stockSymbol"`
	Filename       string  `xml:"filename"`
	Funds          float64 `xml:"funds"`
}

// QuoteServer: Any communication with the quoter server
type QuoteServer struct {
	Timestamp       int     `xml:"timestamp"`
	Server          string  `xml:"server"`
	TransactionNum  int     `xml:"transactionNum"`
	Price           float64 `xml:"price"`
	StockSymbol     string  `xml:"stockSymbol"`
	Username        string  `xml:"username"`
	QuoteServerTime int     `xml:"quoteServerTime"`
	Cryptokey       string  `xml:"cryptokey"`
}

// AccountTransaction: any change in User's account
type AccountTransaction struct {
	Timestamp      int     `xml:"timestamp"`
	Server         string  `xml:"server"`
	TransactionNum int     `xml:"transactionNum"`
	Action         string  `xml:"action"`
	Username       string  `xml:"username"`
	Funds          float64 `xml:"funds"`
}

// SystemEvent: Any event that is triggered by our system. For example, buying a stock because a trigger was set by the user.
type SystemEvent struct {
	Timestamp      int     `xml:"timestamp"`
	Server         string  `xml:"server"`
	TransactionNum int     `xml:"transactionNum"`
	Command        string  `xml:"command"`
	Username       string  `xml:"username"`
	StockSymbol    string  `xml:"stockSymbol"`
	Filename       string  `xml:"filename"`
	Funds          float64 `xml:"funds"`
}

// ErrorEvent: Any error that occurs for a transaction with the quote server
type ErrorEvent struct {
	Timestamp      int     `xml:"timestamp"`
	Server         string  `xml:"server"`
	TransactionNum int     `xml:"transactionNum"`
	Command        string  `xml:"command"`
	Username       string  `xml:"username"`
	StockSymbol    string  `xml:"stockSymbol"`
	Filename       string  `xml:"filename"`
	Funds          float64 `xml:"funds"`
	ErrorMessage   string  `xml:"errorMessage"`
}

// Debug: debug logs for ourselves
type Debug struct {
	Timestamp      int     `xml:"timestamp"`
	Server         string  `xml:"server"`
	TransactionNum int     `xml:"transactionNum"`
	Command        string  `xml:"command"`
	Username       string  `xml:"username"`
	StockSymbol    string  `xml:"stockSymbol"`
	Filename       string  `xml:"filename"`
	Funds          float64 `xml:"funds"`
	DebugMessage   string  `xml:"debugMessage"`
}