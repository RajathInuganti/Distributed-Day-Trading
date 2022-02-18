package main

import (
	"encoding/xml"
	"fmt"
)

type Command struct {
	XMLName  xml.Name `xml:"plant"`
	Timestamp string  `xml:"Timestamp,attr"`
	Command  string   `xml:"Command,attr"`
	Username string   `xml:"Username,attr"`
	Amount   string   `xml:"Amount,attr"`
	Stock    string   `xml:"Stock,attr"`
	Filename string   `xml:"Filename,attr"`
}

func CreateXML() {
	// list := []*Command{
	// 	&Command{Command: "ADD", Username: "KS", Amount: "23"},
	// 	&Command{Command: "BUY", Username: "KS", Stock: "ABS"},
	// }

	c := Command{Command: "ADD", Username: "KS", Amount: "23"}

	out , _ := xml.MarshalIndent(&c, " ", "  ")
	fmt.Println(string(out))


}


func main() {
	CreateXML()
}