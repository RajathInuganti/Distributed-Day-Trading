package logs

import "encoding/xml"

// Log: The created Log object will have any *ONE* of the following fields.
// https://www.ece.uvic.ca/~seng468/ProjectWebSite/logfile.xsd
type Log struct {
	XMLName xml.Name `xml:"log"`
	UserCommand        []*UserCommand        `xml:"userCommand"`
	QuoteServer        []*QuoteServer        `xml:"quoteServer"`
	AccountTransaction []*AccountTransaction `xml:"accountTransaction"`
	SystemEvent        []*SystemEvent        `xml:"systemEvent"`
	ErrorEvent         []*ErrorEvent         `xml:"errorEvent"`
	DebugEvent         []*Debug              `xml:"debugEvent"`
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