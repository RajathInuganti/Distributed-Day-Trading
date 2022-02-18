package logs

import (
	"encoding/xml"
	"fmt"
)

func CreateXML() {
	log1 := &Log{
		UserCommand: []*UserCommand{
			{
				Timestamp:      123,
				Server:         "server1",
				TransactionNum: 1,
				Command:        "buy",
				Username:       "user1",
				StockSymbol:    "AAPL",
			},
		},
	}

	log2 := &Log{
		QuoteServer: []*QuoteServer{
			{
				Timestamp:       124,
				Server:          "server2",
				TransactionNum:  2,
				Price:           100.0,
				StockSymbol:     "AAPL",
				Username:        "user1",
				QuoteServerTime: 123,
				Cryptokey:       "key1",
			},
		},
	}

	log3 := &Log{
		AccountTransaction: []*AccountTransaction{
			{
				Timestamp:      125,
				Server:         "server3",
				TransactionNum: 3,
				Action:         "buy",
				Username:       "user1",
				Funds:          100.0,
			},
		},
	}

	output := xml.Header

	marshal , _ := xml.MarshalIndent(log1, " ", "    ")
	output += string(marshal) + "\n"

	marshal , _ = xml.MarshalIndent(log2, " ", "    ")
	output += string(marshal) + "\n"

	marshal , _ = xml.MarshalIndent(log3, " ", "    ")
	output += string(marshal) + "\n"

	fmt.Println(string(output))

	
}